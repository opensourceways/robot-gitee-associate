package main

import (
	"fmt"

	sdk "gitee.com/openeuler/go-gitee/gitee"
	libconfig "github.com/opensourceways/community-robot-lib/config"
	"github.com/opensourceways/community-robot-lib/giteeclient"
	libplugin "github.com/opensourceways/community-robot-lib/giteeplugin"
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

func (bot *robot) NewPluginConfig() libconfig.PluginConfig {
	return &configuration{}
}

func (bot *robot) getConfig(cfg libconfig.PluginConfig, org, repo string) (*botConfig, error) {
	c, ok := cfg.(*configuration)
	if !ok {
		return nil, fmt.Errorf("can't convert to configuration")
	}

	if bc := c.configFor(org, repo); bc != nil {
		return bc, nil
	}

	return nil, fmt.Errorf("no config for this repo:%s/%s", org, repo)
}

func (bot *robot) RegisterEventHandler(p libplugin.HandlerRegitster) {
	p.RegisterPullRequestHandler(bot.handlePREvent)
	p.RegisterNoteEventHandler(bot.handleNoteEvent)
	p.RegisterIssueHandler(bot.handleIssueEvent)
}

func (bot *robot) handlePREvent(e *sdk.PullRequestEvent, pc libconfig.PluginConfig, log *logrus.Entry) error {
	action := giteeclient.GetPullRequestAction(e)
	if action != giteeclient.PRActionOpened {
		return nil
	}

	org, repo := giteeclient.GetOwnerAndRepoByPREvent(e)

	if _, err := bot.getConfig(pc, org, repo); err != nil {
		return err
	}

	return bot.handlePRCreate(e, log)
}

func (bot *robot) handleNoteEvent(e *sdk.NoteEvent, pc libconfig.PluginConfig, log *logrus.Entry) error {
	switch e.GetNoteableType() {
	case "PullRequest":
		ne := giteeclient.NewPRNoteEvent(e)

		if !ne.IsCreatingCommentEvent() {
			log.Info("Event is not a creation of a comment, skipping.")

			return nil
		}

		return bot.handlePRComment(e)
	case "Issue":
		ne := giteeclient.NewIssueNoteEvent(e)
		if !ne.IsCreatingCommentEvent() {
			log.Info("Event is not a creation of a comment, skipping.")

			return nil
		}

		cfg, err := bot.getConfig(pc, ne.GetRepository().GetNameSpace(), ne.GetRepository().GetPath())
		if err != nil {
			return err
		}

		if cfg.SwitchOfMilestone != "on" {
			return nil
		}

		return bot.handleIssueComment(e)
	}

	return nil
}

func (bot *robot) handleIssueEvent(e *sdk.IssueEvent, pc libconfig.PluginConfig, log *logrus.Entry) error {
	org, repo := giteeclient.GetOwnerAndRepoByIssueEvent(e)

	cfg, err := bot.getConfig(pc, org, repo)
	if err != nil {
		return err
	}

	if e.GetAction() == "open" && cfg.SwitchOfMilestone != "on" {
		return nil
	}

	return bot.handleIssueCreate(e, log)
}
