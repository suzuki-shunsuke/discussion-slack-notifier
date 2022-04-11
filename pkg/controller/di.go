package controller

import (
	"context"

	"github.com/suzuki-shunsuke/discussion-slack-notifier/pkg/github"
	"github.com/suzuki-shunsuke/discussion-slack-notifier/pkg/input"
)

func newGitHub(ctx context.Context, param *input.Param) github.GitHub {
	return github.New(ctx, param.GitHubToken)
}
