package main

import (
	"errors"

	sdk "github.com/google/go-github/v36/github"
	"github.com/opensourceways/server-common-lib/config"
	"github.com/sirupsen/logrus"

	"github.com/opensourceways/robot-github-lib/framework"
)

// TODO: set botName
const botName = ""

type iClient interface {
}

func newRobot(cli iClient) *robot {
	return &robot{cli: cli}
}

type robot struct {
	cli iClient
}

func (bot *robot) NewConfig() config.Config {
	return &configuration{}
}

func (bot *robot) getConfig(cfg config.Config) (*configuration, error) {
	if c, ok := cfg.(*configuration); ok {
		return c, nil
	}
	return nil, errors.New("can't convert to configuration")
}

func (bot *robot) RegisterEventHandler(f framework.HandlerRegister) {
	f.RegisterIssueHandler(bot.handleIssueEvent)
	f.RegisterPullRequestHandler(bot.handlePREvent)
	f.RegisterIssueCommentHandler(bot.handleNoteEvent)
	f.RegisterPushEventHandler(bot.handlePushEvent)
}

func (bot *robot) handlePREvent(e *sdk.PullRequestEvent, c config.Config, log *logrus.Entry) error {
	// TODO: if it doesn't needd to hand PR event, delete this function.
	return nil
}

func (bot *robot) handleIssueEvent(e *sdk.IssuesEvent, c config.Config, log *logrus.Entry) error {
	// TODO: if it doesn't needd to hand Issue event, delete this function.
	return nil
}

func (bot *robot) handlePushEvent(e *sdk.PushEvent, c config.Config, log *logrus.Entry) error {
	// TODO: if it doesn't needd to hand Push event, delete this function.
	return nil
}

func (bot *robot) handleNoteEvent(e *sdk.IssueCommentEvent, c config.Config, log *logrus.Entry) error {
	// TODO: if it doesn't needd to hand Note event, delete this function.
	return nil
}
