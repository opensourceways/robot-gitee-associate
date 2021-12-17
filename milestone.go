package main

import (
	"fmt"
	"regexp"

	sdk "gitee.com/openeuler/go-gitee/gitee"
	"github.com/sirupsen/logrus"
)

const (
	milestoneHasSetMessage = "Milestones have been set when the issue (%s)was created"
	unsetMilestoneLabel    = "needs-milestone"
	unsetMilestoneComment  = "@%s You have not selected a milestone,please select a milestone." +
		"After setting the milestone, " +
		"you can use the **/check-milestone** command to remove the **needs-milestone** label."
)

var checkMilestoneRe = regexp.MustCompile(`(?mi)^/check-milestone\s*$`)

func (bot *robot) handleIssueCreate(e *sdk.IssueEvent, log *logrus.Entry) error {
	if e.GetIssue().Milestone != nil && e.GetIssue().Milestone.Id != 0 {
		log.Debug(fmt.Sprintf(milestoneHasSetMessage, e.GetIssue().GetNumber()))

		return nil
	}

	owner := e.GetRepository().GetNameSpace()
	repo := e.GetRepository().GetPath()
	number := e.GetIssue().GetNumber()
	author := e.GetIssue().GetUser().GetLogin()

	return bot.handleAddIssueLabelAndComment(owner, repo, number, author)
}

func (bot *robot) handleIssueComment(e *sdk.NoteEvent) error {
	if !checkMilestoneRe.MatchString(e.GetComment().GetBody()) {
		return nil
	}

	org, repo := e.GetRepository().GetOwnerAndRepo()
	number := e.GetIssue().GetNumber()
	author := e.GetIssue().GetUser().GetLogin()

	issueHasLabels := e.GetIssue().Labels
	hasMilestone := e.GetIssue().Milestone != nil && e.GetIssue().Milestone.Id != 0

	hasLabel := false

	for _, v := range issueHasLabels {
		if v.Name == unsetMilestoneLabel {
			hasLabel = true
		}
	}

	if hasLabel && hasMilestone {
		return bot.cli.RemoveIssueLabel(org, repo, number, unsetMilestoneLabel)
	}

	if !hasMilestone && !hasLabel {
		return bot.handleAddIssueLabelAndComment(org, repo, number, author)
	}

	return nil
}

func (bot *robot) handleAddIssueLabelAndComment(owner, repo, number, author string) error {
	err := bot.cli.AddIssueLabel(owner, repo, number, unsetMilestoneLabel)
	if err != nil {
		return err
	}

	return bot.cli.CreateIssueComment(owner, repo, number, fmt.Sprintf(unsetMilestoneComment, author))
}
