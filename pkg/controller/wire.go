//go:build wireinject
// +build wireinject

package controller

import (
	"context"

	"github.com/google/wire"
	"github.com/suzuki-shunsuke/discussion-slack-notifier/pkg/config"
	filter "github.com/suzuki-shunsuke/discussion-slack-notifier/pkg/entry-filter"
	"github.com/suzuki-shunsuke/discussion-slack-notifier/pkg/input"
)

func InitializeController(ctx context.Context, param *input.Param) *Controller {
	wire.Build(New, newGitHub, newSlack, config.NewReader, newPayloadReader, filter.New)
	return &Controller{}
}
