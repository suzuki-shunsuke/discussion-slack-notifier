package controller

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/go-github/v43/github"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/suzuki-shunsuke/discussion-slack-notifier/pkg/config"
	filter "github.com/suzuki-shunsuke/discussion-slack-notifier/pkg/entry-filter"
	gh "github.com/suzuki-shunsuke/discussion-slack-notifier/pkg/github"
	"github.com/suzuki-shunsuke/discussion-slack-notifier/pkg/input"
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

	slackChannelNames := ctrl.listTargetChannels(ctx, cfg, payload, labels)

	if slackChannelNames.Len() == 0 {
		logrus.Info("No notification is sent")
		return nil
	}
	logrus.WithField("channels", slackChannelNames.String()).Info("notified channels")

	chMap, err := ctrl.listAllChannels(ctx, cfg)
	if err != nil {
		return err
	}
	channelIDs := ctrl.listChannelIDs(slackChannelNames, chMap)
	msg, err := ctrl.getMessage(payload)
	if err != nil {
		return err
	}
	if err := ctrl.notifyChannels(ctx, channelIDs, slack.MsgOptionText(msg, false)); err != nil {
		return err
	}
	return nil
}

func (ctrl *Controller) getMessage(payload *github.DiscussionEvent) (string, error) { //nolint:unparam
	discussion := payload.GetDiscussion()
	txt := fmt.Sprintf(`# %s

Category: %s`, discussion.GetTitle(), discussion.GetDiscussionCategory().GetName())
	return txt, nil
}

func (ctrl *Controller) readConfig(p string, cfg *config.Config) error {
	return ctrl.cfgReader.Read(p, cfg) //nolint:wrapcheck
}

func (ctrl *Controller) listChannelIDs(chNames *util.StrSet, chMap map[string]string) *util.StrSet {
	channelIDs := util.NewStrSet(chNames.Len())
	for channelName := range chNames.Map() {
		chID, ok := chMap[channelName]
		if !ok {
			// channelName is invalid
			continue
		}
		channelIDs.Add(chID)
	}
	return channelIDs
}

func (ctrl *Controller) readPayload(p string, payload *github.DiscussionEvent) error {
	return ctrl.payloadReader.Read(p, payload) //nolint:wrapcheck
}

func (ctrl *Controller) listLabels(ctx context.Context, owner, repo string, discussID int) (*util.StrSet, error) {
	return ctrl.github.ListDiscussionLabels(ctx, owner, repo, discussID) //nolint:wrapcheck
}

func (ctrl *Controller) listTargetChannels(ctx context.Context, cfg *config.Config, payload *github.DiscussionEvent, labels *util.StrSet) *util.StrSet {
	channels := util.NewStrSet(0)
	for _, entry := range cfg.Entries {
		f, err := ctrl.filterEntry(ctx, entry, cfg, payload, labels)
		if err != nil {
			logrus.WithError(err).Error("filter an entry")
		}
		if f {
			channels.Append(entry.Channels...)
		}
	}
	return channels
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

func (ctrl *Controller) notifyChannels(ctx context.Context, slackChannels *util.StrSet, opts ...slack.MsgOption) error {
	var oneErr error
	for slackChannel := range slackChannels.Map() {
		// notify to slack
		if err := ctrl.notify(ctx, slackChannel, opts...); err != nil {
			oneErr = err
			logrus.WithError(err).Error("notify to slack")
		}
	}
	return oneErr
}
