package matcher

import (
	"errors"

	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/util"
	"github.com/innogames/slack-bot/client"
	"github.com/slack-go/slack"
)

// NewAdminMatcher is a wrapper to only executable by a whitelisted admin user
func NewAdminMatcher(cfg config.Config, slackClient client.SlackClient, matcher Matcher) Matcher {
	return adminMatcher{matcher, cfg, slackClient}
}

type adminMatcher struct {
	matcher     Matcher
	cfg         config.Config
	slackClient client.SlackClient
}

func (m adminMatcher) Match(event slack.MessageEvent) (Runner, Result) {
	run, result := m.matcher.Match(event)
	if !result.Matched() {
		return nil, result
	}

	for _, adminId := range m.cfg.AdminUsers {
		if adminId == event.User {
			return run, result
		}
	}

	match := MapResult{
		util.FullMatch: event.Text,
	}

	return func(match Result, event slack.MessageEvent) {
		m.slackClient.ReplyError(
			event,
			errors.New("Sorry, you are no admin and not allowed to execute this command!"),
		)
	}, match
}
