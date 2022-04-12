package controller

import (
	"context"

	"github.com/google/go-github/v43/github"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/discussion-slack-notifier/pkg/config"
	"github.com/suzuki-shunsuke/discussion-slack-notifier/pkg/util"
)

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
