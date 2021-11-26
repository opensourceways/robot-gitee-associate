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
	ne := giteeclient.NewPRNoteEvent(e)
	pr := ne.GetPRInfo()

	if !ne.IsCreatingCommentEvent() {
		log.Info("Event is not a creation of a comment, skipping.")

		return nil
	}

	if _, err := bot.getConfig(pc, pr.Org, pr.Repo); err != nil {
		return err
	}

	if ne.IsPullRequest() {
		return bot.handlePRComment(e)
	}

	return nil
}
