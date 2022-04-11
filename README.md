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

## GitHub Token's permission

GitHub Token is required to list Discussion's labels.
You can use GitHub Actions' token `github.token`.

* discussions: read

## Environment Variables

* GITHUB_TOKEN: GitHub Access Token. This is required to list the Discussion's labels
* SLACK_TOKEN: 
* GITHUB_EVENT_PATH: 

## Configuration

e.g.

```yaml
entries:
- labels:
  - foo
  channels:
  - general
channels:
  general: XXXXXXXXX
```

In case of above configuration, if the Discussion Label `foo` is set, the notification is sent to the slack channel `general`.

## LICENSE

[MIT](LICENSE)
