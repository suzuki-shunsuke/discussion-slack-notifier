package github

import (
	"context"
	"fmt"

	"github.com/shurcooL/githubv4"
	"github.com/suzuki-shunsuke/discussion-slack-notifier/pkg/util"
	"golang.org/x/oauth2"
)

type githubClient struct {
	v4Client *githubv4.Client
}

type GitHub interface {
	ListDiscussionLabels(ctx context.Context, owner, repo string, discussID int) (*util.StrSet, error)
}

func New(ctx context.Context, token string) GitHub {
	return &githubClient{
		v4Client: githubv4.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		))),
	}
}

type Label struct {
	Name string `json:"name"`
}

func (client *githubClient) ListDiscussionLabels(ctx context.Context, owner, repo string, discussionID int) (*util.StrSet, error) {
	// https://github.com/shurcooL/githubv4#pagination
	var q struct {
		Repository struct {
			Discussion struct {
				Labels struct {
					Nodes    []*Label
					PageInfo struct {
						EndCursor   githubv4.String
						HasNextPage bool
					}
				} `graphql:"labels(first: 100, after: $labelsCursor)"` // 100 per page.
			} `graphql:"discussion(number: $discussionNumber)"`
		} `graphql:"repository(owner: $repositoryOwner, name: $repositoryName)"`
	}
	variables := map[string]interface{}{
		"repositoryOwner":  githubv4.String(owner),
		"repositoryName":   githubv4.String(repo),
		"discussionNumber": githubv4.Int(discussionID),
		"labelsCursor":     (*githubv4.String)(nil), // Null after argument to get first page.
	}

	var labels []*Label
	for {
		if err := client.v4Client.Query(ctx, &q, variables); err != nil {
			return nil, fmt.Errorf("list issue comments by GitHub API: %w", err)
		}
		labels = append(labels, q.Repository.Discussion.Labels.Nodes...)
		if !q.Repository.Discussion.Labels.PageInfo.HasNextPage {
			break
		}
		variables["labelsCursor"] = githubv4.NewString(q.Repository.Discussion.Labels.PageInfo.EndCursor)
	}
	arr := util.NewStrSet(len(labels))
	for _, label := range labels {
		arr.Add(label.Name)
	}
	return arr, nil
}
