package main

import (
	"fmt"
	"regexp"

	"github.com/opensourceways/community-robot-lib/giteeclient"
	sdk "github.com/opensourceways/go-gitee/gitee"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/sets"
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
	return bot.handlePRIssue(org, repo, e.GetPRAuthor(), e.GetPRNumber(), e.GetPRLabelSet())
}

func (bot *robot) handlePRComment(e *sdk.NoteEvent) error {
	c := e.GetComment().GetBody()

	if checkIssueRe.MatchString(c) {
		return bot.handleCheckIssue(e)
	}

	if removeMissIssue.MatchString(c) {
		return bot.handleRemoveMissLabel(e)
	}

	return nil
}

func (bot *robot) handleCheckIssue(e *sdk.NoteEvent) error {
	org, repo := e.GetOrgRepo()
	return bot.handlePRIssue(org, repo, e.GetPRAuthor(), e.GetPRNumber(), e.GetPRLabelSet())
}

func (bot *robot) handlePRIssue(org, repo, prAuthor string, number int32, labels sets.String) error {
	issues, err := bot.cli.ListPrIssues(org, repo, number)
	if err != nil {
		return err
	}

	hasIssue := len(issues) > 0
	hasLabel := labels.Has(missIssueLabel)

	if !hasIssue && !hasLabel {
		if err := bot.cli.AddPRLabel(org, repo, number, missIssueLabel); err != nil {
			return err
		}

		return bot.cli.CreatePRComment(org, repo, number, fmt.Sprintf(missIssueComment, prAuthor))
	}

	if hasIssue && hasLabel {
		return bot.cli.RemovePRLabel(org, repo, number, missIssueLabel)
	}

	return nil
}

func (bot *robot) handleRemoveMissLabel(e *sdk.NoteEvent) error {
	if !e.GetPRLabelSet().Has(missIssueLabel) {
		return nil
	}

	org, repo := e.GetOrgRepo()

	b, err := bot.cli.IsCollaborator(org, repo, e.GetCommenter())
	if err != nil {
		return err
	}

	number := e.GetPRNumber()

	if !b {
		msg := "Only members of the repository can delete the 'needs-issue' label. Please contact them to do it."

		return bot.cli.CreatePRComment(
			org, repo, number,
			giteeclient.GenResponseWithReference(e, msg),
		)
	}

	return bot.cli.RemovePRLabel(org, repo, number, missIssueLabel)
}
