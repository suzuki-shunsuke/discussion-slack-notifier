package controller

import (
	"context"

	"github.com/slack-go/slack"
	"github.com/suzuki-shunsuke/discussion-slack-notifier/pkg/input"
)

type Slack interface {
	PostMessageContext(ctx context.Context, channelID string, options ...slack.MsgOption) (string, string, error)
}

func newSlack(param *input.Param) Slack {
	return slack.New(param.SlackToken)
}
