# discussion-slack-notifier

Notify GitHub Discussions events to Slack with GitHub Actions.

## :warning: Development Status

This is still in alpha. This isn't production ready yet.

## Motivation

You can receive the notification about GitHub Discussions to Slack with the official GitHub App.

https://github.com/github/feedback/discussions/2844

e.g.

<img width="417" alt="image" src="https://user-images.githubusercontent.com/13323303/162709503-90875f17-8879-45e1-b47c-3bb59af20847.png">

This integration is very useful, but you can't configure the notification flexibly.

`discussion-slack-notifier` supports configuring the notification flexibly.

* Change notification channels according to the Discussion Labels

## Feature

* Change notification channels according to the Discussion Labels

## Requirement

* Slack App Bot Token

Please create a Slack App.

## Slack App's permission

* chat:write
* chat:write.public (Optional)

## How to use

Please set up GitHub Actions Workflow to run discussion-slack-notifier.

1. Add GitHub Actions Secret `SLACK_TOKEN`
1. Add configuration file [discussion-slack-notifier.yaml](discussion-slack-notifier.yaml)
1. Add GitHub Actions workflow. e.g. [notify-discuss.yaml](.github/workflows/notify-discuss.yaml)

## GitHub Token's permission

GitHub Token is required to list Discussion's labels.
You can use GitHub Actions' token `github.token`.

* discussions: read

## Environment Variables

* GITHUB_TOKEN: GitHub Access Token. This is required to list the Discussion's labels
* SLACK_TOKEN: 
* GITHUB_EVENT_PATH: 

## Configuration

e.g. [discussion-slack-notifier.yaml](discussion-slack-notifier.yaml)

### Templates

You can customize notification message.

* [Go's text/template](https://pkg.go.dev/text/template)
* http://masterminds.github.io/sprig/

#### Template Priority

1. entry's `template`
1. entry's `template_name`
1. `templates`'s `default` template
1. Built in default template

#### Template Variables

* Title (string): Discussion title
* CategoryName (string): Discussion Category Name
* Vars (map[string]interface{}): user defined variables

## LICENSE

[MIT](LICENSE)
