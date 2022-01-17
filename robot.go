package main

import (
	"fmt"

	"github.com/opensourceways/community-robot-lib/config"
	"github.com/opensourceways/community-robot-lib/robot-gitee-framework"
	sdk "github.com/opensourceways/go-gitee/gitee"
	"github.com/sirupsen/logrus"
)

const botName = "associate"

type iClient interface {
	AddPRLabel(org, repo string, number int32, label string) error
	RemovePRLabel(org, repo string, number int32, label string) error
	CreatePRComment(org, repo string, number int32, comment string) error
	IsCollaborator(owner, repo, login string) (bool, error)
	ListPrIssues(org, repo string, number int32) ([]sdk.Issue, error)
	CreateIssueComment(org, repo string, number string, comment string) error
	RemoveIssueLabel(org, repo, number, label string) error
	AddIssueLabel(org, repo, number, label string) error
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

func (bot *robot) getConfig(cfg config.Config, org, repo string) (*botConfig, error) {
	c, ok := cfg.(*configuration)
	if !ok {
		return nil, fmt.Errorf("can't convert to configuration")
	}

	if bc := c.configFor(org, repo); bc != nil {
		return bc, nil
	}

	return nil, fmt.Errorf("no config for this repo:%s/%s", org, repo)
}

func (bot *robot) RegisterEventHandler(f framework.HandlerRegitster) {
	f.RegisterPullRequestHandler(bot.handlePREvent)
	f.RegisterNoteEventHandler(bot.handleNoteEvent)
	f.RegisterIssueHandler(bot.handleIssueEvent)
}

func (bot *robot) handlePREvent(e *sdk.PullRequestEvent, pc config.Config, log *logrus.Entry) error {
	action := sdk.GetPullRequestAction(e)
	if action != sdk.PRActionOpened && action != sdk.PRActionLinkIssue {
		return nil
	}

	org, repo := e.GetOrgRepo()

	cfg, err := bot.getConfig(pc, org, repo)
	if err != nil {
		return err
	}

	if !cfg.enableCheckingIssue() {
		return nil
	}

	return bot.handlePRIssue(org, repo, e.GetPullRequest())
}

func (bot *robot) handleNoteEvent(e *sdk.NoteEvent, pc config.Config, log *logrus.Entry) error {
	if !e.IsCreatingCommentEvent() {
		return nil
	}

	org, repo := e.GetOrgRepo()

	cfg, err := bot.getConfig(pc, org, repo)
	if err != nil {
		return err
	}

	if e.IsPullRequest() {
		return bot.handlePRComment(e, cfg)
	}

	if e.IsIssue() {
		return bot.handleIssueComment(e, cfg)
	}

	return nil
}

func (bot *robot) handleIssueEvent(e *sdk.IssueEvent, pc config.Config, log *logrus.Entry) error {
	if e.GetAction() != sdk.ActionOpen {
		return nil
	}

	org, repo := e.GetOrgRepo()

	cfg, err := bot.getConfig(pc, org, repo)
	if err != nil {
		return err
	}

	if !cfg.enableCheckingMilestone() {
		return nil
	}

	return bot.handleIssueCreate(e, log)
}
