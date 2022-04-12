package controller

import (
	"errors"
	"fmt"

	"github.com/google/go-github/v43/github"
	"github.com/suzuki-shunsuke/discussion-slack-notifier/pkg/config"
	"github.com/suzuki-shunsuke/discussion-slack-notifier/pkg/template"
)

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

func getTemplateParam(payload *github.DiscussionEvent, cfg *config.Config, entry *config.Entry) interface{} {
	discussion := payload.GetDiscussion()
	return map[string]interface{}{
		"Title":        discussion.GetTitle(),
		"CategoryName": discussion.GetDiscussionCategory().GetName(),
	}
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
	txt, err := template.Execute(tpl, getTemplateParam(payload, cfg, entry))
	if err != nil {
		return "", fmt.Errorf("render a message template: %w", err)
	}
	return txt, nil
}
