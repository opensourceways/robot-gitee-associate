package main

import (
	"fmt"
	"regexp"

	sdk "gitee.com/openeuler/go-gitee/gitee"
	"github.com/opensourceways/community-robot-lib/giteeclient"
	"github.com/sirupsen/logrus"
)

const (
	missIssueComment = "@%s PullRequest must be associated with an issue, please associate at least one issue. " +
		"after associating an issue, you can use the **/check-issue** command to remove the **needs-issue** label."
	missIssueLabel = "needs-issue"
)

var (
	checkIssueRe    = regexp.MustCompile(`(?mi)^/check-issue\s*$`)
	removeMissIssue = regexp.MustCompile(`(?mi)^/remove-needs-issue\s*$`)
)

func (bot *robot) handlePRCreate(e *sdk.PullRequestEvent, log *logrus.Entry) error {
	prInfo := giteeclient.GetPRInfoByPREvent(e)

	issues, err := bot.cli.ListPrIssues(prInfo.Org, prInfo.Repo, prInfo.Number)
	if err != nil {
		log.Errorf("get issues of pr failed: %v", err)

		return err
	}

	hasLabel := prInfo.HasLabel(missIssueLabel)

	if len(issues) == 0 && !hasLabel {
		err = bot.cli.AddPRLabel(prInfo.Org, prInfo.Repo, prInfo.Number, missIssueLabel)
		if err != nil {
			return err
		}

		return bot.cli.CreatePRComment(prInfo.Org, prInfo.Repo, prInfo.Number,
			fmt.Sprintf(missIssueComment, prInfo.Author))
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
	ne := giteeclient.NewPRNoteEvent(e)
	pr := ne.GetPRInfo()

	issues, err := bot.cli.ListPrIssues(pr.Org, pr.Repo, pr.Number)
	if err != nil {
		return err
	}

	hasLabel := pr.HasLabel(missIssueLabel)

	if len(issues) == 0 && !hasLabel {
		if err := bot.cli.AddPRLabel(pr.Org, pr.Repo, pr.Number, missIssueLabel); err != nil {
			return err
		}

		return bot.cli.CreatePRComment(pr.Org, pr.Repo, pr.Number, fmt.Sprintf(missIssueComment, pr.Author))
	}

	if len(issues) > 0 && hasLabel {
		return bot.cli.RemovePRLabel(pr.Org, pr.Repo, pr.Number, missIssueLabel)
	}

	return nil
}

func (bot *robot) handleRemoveMissLabel(e *sdk.NoteEvent) error {
	ne := giteeclient.NewPRNoteEvent(e)
	pr := ne.GetPRInfo()

	isCo, err := bot.cli.IsCollaborator(pr.Org, pr.Repo, ne.GetCommenter())
	if err != nil {
		return err
	}

	if !isCo {
		comment := fmt.Sprintf("@%s Members of the repository can delete the 'needs-issue' label. "+
			"Please contact the Members.", ne.GetCommenter())

		return bot.cli.CreatePRComment(pr.Org, pr.Repo, pr.Number, comment)
	}

	return bot.cli.RemovePRLabel(pr.Org, pr.Repo, pr.Number, missIssueLabel)
}
