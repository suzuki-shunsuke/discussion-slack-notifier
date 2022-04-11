package filter

import (
	"context"

	"github.com/google/go-github/v43/github"
	"github.com/suzuki-shunsuke/discussion-slack-notifier/pkg/config"
	"github.com/suzuki-shunsuke/discussion-slack-notifier/pkg/util"
)

type Filter interface {
	Filter(ctx context.Context, entry *config.Entry, cfg *config.Config, payload *github.DiscussionEvent, labels *util.StrSet) (bool, error)
}

type entryFilter struct {
	filters []Filter
}

func New() Filter {
	return &entryFilter{
		filters: []Filter{
			&labelEntryFilter{},
		},
	}
}

func (filter *entryFilter) Filter(ctx context.Context, entry *config.Entry, cfg *config.Config, payload *github.DiscussionEvent, labels *util.StrSet) (bool, error) {
	for _, flt := range filter.filters {
		f, err := flt.Filter(ctx, entry, cfg, payload, labels)
		if err != nil {
			return false, err
		}
		if !f {
			return false, nil
		}
	}
	return true, nil
}

type labelEntryFilter struct{}

func (filter *labelEntryFilter) Filter(ctx context.Context, entry *config.Entry, cfg *config.Config, payload *github.DiscussionEvent, labels *util.StrSet) (bool, error) {
	if len(entry.Labels) == 0 {
		return true, nil
	}
	for _, label := range entry.Labels {
		if labels.Has(label) {
			return true, nil
		}
	}
	return false, nil
}
