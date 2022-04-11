package cli

import (
	"context"
	"io"
	"os"

	"github.com/suzuki-shunsuke/discussion-slack-notifier/pkg/controller"
	"github.com/suzuki-shunsuke/discussion-slack-notifier/pkg/input"
)

type Runner struct {
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
	LDFlags *LDFlags
}

type LDFlags struct {
	Version string
	Commit  string
	Date    string
}

func (runner *Runner) Run(ctx context.Context, args ...string) error {
	param := &input.Param{
		GitHubToken: os.Getenv("GITHUB_TOKEN"),
		SlackToken:  os.Getenv("SLACK_TOKEN"),
		PayloadPath: os.Getenv("GITHUB_EVENT_PATH"),
		ConfigPath:  "discussion-slack-notifier.yaml",
	}
	ctrl := controller.InitializeController(ctx, param)
	return ctrl.Run(ctx, param) //nolint:wrapcheck
}
