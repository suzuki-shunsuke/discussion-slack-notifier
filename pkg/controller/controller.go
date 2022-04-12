package controller

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/go-github/v43/github"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/suzuki-shunsuke/discussion-slack-notifier/pkg/config"
	filter "github.com/suzuki-shunsuke/discussion-slack-notifier/pkg/entry-filter"
	gh "github.com/suzuki-shunsuke/discussion-slack-notifier/pkg/github"
	"github.com/suzuki-shunsuke/discussion-slack-notifier/pkg/input"
	"github.com/suzuki-shunsuke/discussion-slack-notifier/pkg/template"
	"github.com/suzuki-shunsuke/discussion-slack-notifier/pkg/util"
)

type PayloadReader interface {
	Read(p string, event *github.DiscussionEvent) error
}

type Controller struct {
	github        gh.GitHub
	slack         Slack
	cfgReader     config.Reader
	payloadReader PayloadReader
	entryFilter   filter.Filter
}

func New(ghClient gh.GitHub, slack Slack, cfgReader config.Reader, payloadReader PayloadReader, entryFilter filter.Filter) *Controller {
	return &Controller{
		github:        ghClient,
		slack:         slack,
		cfgReader:     cfgReader,
		payloadReader: payloadReader,
		entryFilter:   entryFilter,
	}
}

func (ctrl *Controller) Run(ctx context.Context, param *input.Param) error {
	if param.ConfigPath == "" {
		return errors.New("configuration file isn't found")
	}
	if param.PayloadPath == "" {
		return errors.New("GITHUB_EVENT_PATH isn't set")
	}

	cfg := &config.Config{}
	if err := ctrl.readConfig(param.ConfigPath, cfg); err != nil {
		return err
	}

	payload := &github.DiscussionEvent{}
	if err := ctrl.readPayload(param.PayloadPath, payload); err != nil {
		return err
	}

	repo := payload.GetRepo()
	owner := repo.GetOwner()
	discussion := payload.GetDiscussion()
	labels, err := ctrl.listLabels(ctx, owner.GetLogin(), repo.GetName(), discussion.GetNumber())
	if err != nil {
		return err
	}

	entries := ctrl.listTargetEntries(ctx, cfg, payload, labels)

	if len(entries) == 0 {
		logrus.Info("No notification is sent")
		return nil
	}
	logrus.WithField("channels", strings.Join(getChannelNamesFromEntries(entries), ", ")).Info("notified channels")

	chMap, err := ctrl.listAllChannels(ctx, cfg)
	if err != nil {
		return err
	}

	return ctrl.notifies(ctx, cfg, payload, entries, chMap)
}

func getChannelNamesFromEntries(entries map[string]*config.Entry) []string {
	chNames := make([]string, 0, len(entries))
	for k := range entries {
		chNames = append(chNames, k)
	}
	return chNames
}

const defaultMessageTemplate = `# {{.Title}}

Category: {{.CategoryName}}`

func (ctrl *Controller) getMessageTemplate(payload *github.DiscussionEvent, cfg *config.Config, entry *config.Entry) (string, error) { //nolint:unparam
	if entry.Template != "" {
		return entry.Template, nil
	}
	if cfg.Templates == nil {
		return defaultMessageTemplate, nil
	}
	if entry.TemplateName != "" {
		tpl, ok := cfg.Templates[entry.TemplateName]
		if !ok {
			return "", errors.New("template isn't found: " + entry.TemplateName)
		}
		return tpl, nil
	}
	tpl, ok := cfg.Templates["default"]
	if ok {
		return tpl, nil
	}
	return defaultMessageTemplate, nil
}

func (ctrl *Controller) getMessage(payload *github.DiscussionEvent, cfg *config.Config, entry *config.Entry) (string, error) {
	t, err := ctrl.getMessageTemplate(payload, cfg, entry)
	if err != nil {
		return "", err
	}
	tpl, err := template.Parse(t)
	if err != nil {
		return "", fmt.Errorf("parse a message template: %w", err)
	}
	discussion := payload.GetDiscussion()
	txt, err := template.Execute(tpl, map[string]interface{}{
		"Title":        discussion.GetTitle(),
		"CategoryName": discussion.GetDiscussionCategory().GetName(),
	})
	if err != nil {
		return "", fmt.Errorf("render a message template: %w", err)
	}
	return txt, nil
}

func (ctrl *Controller) readConfig(p string, cfg *config.Config) error {
	return ctrl.cfgReader.Read(p, cfg) //nolint:wrapcheck
}

func (ctrl *Controller) readPayload(p string, payload *github.DiscussionEvent) error {
	return ctrl.payloadReader.Read(p, payload) //nolint:wrapcheck
}

func (ctrl *Controller) listLabels(ctx context.Context, owner, repo string, discussID int) (*util.StrSet, error) {
	return ctrl.github.ListDiscussionLabels(ctx, owner, repo, discussID) //nolint:wrapcheck
}

func (ctrl *Controller) listTargetEntries(ctx context.Context, cfg *config.Config, payload *github.DiscussionEvent, labels *util.StrSet) map[string]*config.Entry {
	entries := map[string]*config.Entry{}
	for _, entry := range cfg.Entries {
		f, err := ctrl.filterEntry(ctx, entry, cfg, payload, labels)
		if err != nil {
			logrus.WithError(err).Error("filter an entry")
		}
		if f {
			for _, ch := range entry.Channels {
				if _, ok := entries[ch]; !ok {
					entries[ch] = entry
				}
			}
		}
	}
	return entries
}

func (ctrl *Controller) filterEntry(ctx context.Context, entry *config.Entry, cfg *config.Config, payload *github.DiscussionEvent, labels *util.StrSet) (bool, error) {
	return ctrl.entryFilter.Filter(ctx, entry, cfg, payload, labels) //nolint:wrapcheck
}

func (ctrl *Controller) listAllChannels(ctx context.Context, cfg *config.Config) (map[string]string, error) { //nolint:unparam
	return cfg.Channels, nil
}

func (ctrl *Controller) notify(ctx context.Context, slackChannel string, opts ...slack.MsgOption) error {
	_, _, err := ctrl.slack.PostMessageContext(ctx, slackChannel, opts...)
	if err != nil {
		return fmt.Errorf("post a message to Slack: %w", err)
	}
	return nil
}

func (ctrl *Controller) notifies(ctx context.Context, cfg *config.Config, payload *github.DiscussionEvent, entries map[string]*config.Entry, chMap map[string]string) error {
	var oneErr error
	for chName, entry := range entries {
		chID, ok := chMap[chName]
		if !ok {
			logrus.WithField("channel_name", chName).Error("the channel isn't found")
			continue
		}
		msg, err := ctrl.getMessage(payload, cfg, entry)
		if err != nil {
			logrus.WithField("channel_name", chName).WithError(err).Error("get the message")
			continue
		}
		if err := ctrl.notify(ctx, chID, slack.MsgOptionText(msg, false)); err != nil {
			oneErr = err
			logrus.WithError(err).Error("notify to slack")
		}
	}
	return oneErr
}
