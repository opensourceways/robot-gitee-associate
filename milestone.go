package main

import (
	"fmt"
	"regexp"

	sdk "github.com/opensourceways/go-gitee/gitee"
	"github.com/sirupsen/logrus"
)

const (
	unsetMilestoneLabel = "needs-milestone"

	unsetMilestoneComment = `
@%s , Please select a milestone for the issue. Then, you can use the **/check-milestone** command to remove the **needs-milestone** label.
`
)

var checkMilestoneRe = regexp.MustCompile(`(?mi)^/check-milestone\s*$`)

func (bot *robot) handleIssueCreate(e *sdk.IssueEvent, log *logrus.Entry) error {
	if hasMilestoneOnIssue(e.GetIssue()) {
		return nil
	}

	org, repo := e.GetOrgRepo()

	return bot.handleAddIssueLabelAndComment(org, repo, e.GetIssueNumber(), e.GetIssueAuthor())
}

func (bot *robot) handleIssueComment(e *sdk.NoteEvent, cfg *botConfig) error {
	if !cfg.enableCheckingMilestone() || !checkMilestoneRe.MatchString(e.GetComment().GetBody()) {
		return nil
	}

	org, repo := e.GetOrgRepo()
	number := e.GetIssueNumber()
	hasMilestone := hasMilestoneOnIssue(e.GetIssue())
	hasLabel := e.GetIssueLabelSet().Has(unsetMilestoneLabel)

	if hasMilestone && hasLabel {
		return bot.cli.RemoveIssueLabel(org, repo, number, unsetMilestoneLabel)
	}

	if !hasMilestone && !hasLabel {
		return bot.handleAddIssueLabelAndComment(org, repo, number, e.GetIssueAuthor())
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

func hasMilestoneOnIssue(issue *sdk.IssueHook) bool {
	return issue.GetMilestone().GetID() != 0
}
