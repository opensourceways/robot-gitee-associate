package main

import (
	"fmt"
	"regexp"

	"github.com/opensourceways/community-robot-lib/giteeclient"
	sdk "github.com/opensourceways/go-gitee/gitee"
	"github.com/sirupsen/logrus"
)

const (
	missIssueComment = "@%s PullRequest must be associated with an issue, please associate at least one issue. " +
		"Then, you can use the **/check-issue** command to remove the **needs-issue** label."

	missIssueLabel = "needs-issue"
)

var (
	checkIssueRe    = regexp.MustCompile(`(?mi)^/check-issue\s*$`)
	removeMissIssue = regexp.MustCompile(`(?mi)^/remove-needs-issue\s*$`)
)

func (bot *robot) handlePRCreate(e *sdk.PullRequestEvent, log *logrus.Entry) error {
	org, repo := e.GetOrgRepo()
	number := e.GetPRNumber()

	issues, err := bot.cli.ListPrIssues(org, repo, number)
	if err != nil {
		return err
	}

	hasLabel := e.GetPRLabelSet().Has(missIssueLabel)

	if len(issues) == 0 && !hasLabel {
		err = bot.cli.AddPRLabel(org, repo, number, missIssueLabel)
		if err != nil {
			return err
		}

		return bot.cli.CreatePRComment(
			org, repo, number,
			fmt.Sprintf(missIssueComment, e.GetPRAuthor()),
		)
	}

	return nil
}

func (bot *robot) handlePRComment(e *sdk.NoteEvent) error {
	ne := giteeclient.NewPRNoteEvent(e)

	if checkIssueRe.MatchString(ne.GetComment()) {
		return bot.handleCheckIssue(e)
	}

	if removeMissIssue.MatchString(ne.GetComment()) {
		return bot.handleRemoveMissLabel(e)
	}

	return nil
}

func (bot *robot) handleCheckIssue(e *sdk.NoteEvent) error {
	org, repo := e.GetOrgRepo()
	number := e.GetPRNumber()

	issues, err := bot.cli.ListPrIssues(org, repo, number)
	if err != nil {
		return err
	}

	hasLabel := e.GetPRLabelSets().Has(missIssueLabel)

	if len(issues) == 0 && !hasLabel {
		if err := bot.cli.AddPRLabel(org, repo, number, missIssueLabel); err != nil {
			return err
		}

		return bot.cli.CreatePRComment(org, repo, number, fmt.Sprintf(missIssueComment, e.GetPRAuthor()))
	}

	if len(issues) > 0 && hasLabel {
		return bot.cli.RemovePRLabel(org, repo, number, missIssueLabel)
	}

	return nil
}

func (bot *robot) handleRemoveMissLabel(e *sdk.NoteEvent) error {
	org, repo := e.GetOrgRepo()

	b, err := bot.cli.IsCollaborator(org, repo, e.GetCommenter())
	if err != nil {
		return err
	}

	number := e.GetPRNumber()

	if !b {
		comment := fmt.Sprintf("@%s Members of the repository can delete the 'needs-issue' label. "+
			"Please contact the Members.", e.GetCommenter())

		return bot.cli.CreatePRComment(org, repo, number, comment)
	}

	return bot.cli.RemovePRLabel(org, repo, number, missIssueLabel)
}
