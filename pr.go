package main

import (
	"fmt"
	"regexp"

	"github.com/opensourceways/community-robot-lib/giteeclient"
	sdk "github.com/opensourceways/go-gitee/gitee"
)

const (
	missIssueLabel = "needs-issue"

	missIssueComment = "@%s , PullRequest must be associated with at least one issue."
)

var (
	checkIssueRe    = regexp.MustCompile(`(?mi)^/check-issue\s*$`)
	removeMissIssue = regexp.MustCompile(`(?mi)^/remove-needs-issue\s*$`)
)

func (bot *robot) handlePRComment(e *sdk.NoteEvent, cfg *botConfig) error {
	if !cfg.enableCheckingIssue() {
		return nil
	}

	c := e.GetComment().GetBody()

	if checkIssueRe.MatchString(c) {
		org, repo := e.GetOrgRepo()

		return bot.handlePRIssue(org, repo, e.GetPullRequest())
	}

	if removeMissIssue.MatchString(c) {
		return bot.handleRemoveMissLabel(e)
	}

	return nil
}

func (bot *robot) handlePRIssue(org, repo string, pr *sdk.PullRequestHook) error {
	number := pr.GetNumber()
	labels := pr.LabelsToSet()
	prAuthor := pr.GetUser().GetLogin()
	issues := pr.GetIssues()

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
