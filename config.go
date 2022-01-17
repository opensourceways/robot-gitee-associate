package main

import "github.com/opensourceways/community-robot-lib/config"

type configuration struct {
	ConfigItems []botConfig `json:"config_items,omitempty"`
}

func (c *configuration) configFor(org, repo string) *botConfig {
	if c == nil {
		return nil
	}

	items := c.ConfigItems
	v := make([]config.IRepoFilter, len(items))
	for i := range items {
		v[i] = &items[i]
	}

	if i := config.Find(org, repo, v); i >= 0 {
		return &items[i]
	}

	return nil
}

func (c *configuration) Validate() error {
	if c == nil {
		return nil
	}

	items := c.ConfigItems
	for i := range items {
		if err := items[i].validate(); err != nil {
			return err
		}
	}

	return nil
}

func (c *configuration) SetDefault() {
	if c == nil {
		return
	}

	Items := c.ConfigItems
	for i := range Items {
		Items[i].setDefault()
	}
}

type botConfig struct {
	config.RepoFilter

	// EnableCheckAssociateIssue Controls whether to check PR related issues, default true.
	EnableCheckAssociateIssue *bool `json:"enable_check_associate_issue,omitempty"`

	// EnableCheckAssociateMilestone Controls whether to check issue-related milestones, default true
	EnableCheckAssociateMilestone *bool `json:"enable_check_associate_milestone,omitempty"`
}

func (c *botConfig) setDefault() {
}

func (c *botConfig) validate() error {
	return c.RepoFilter.Validate()
}

func (c *botConfig) enableCheckingIssue() bool {
	if c.EnableCheckAssociateIssue == nil {
		return true
	}

	return *c.EnableCheckAssociateIssue
}

func (c *botConfig) enableCheckingMilestone() bool {
	if c.EnableCheckAssociateMilestone == nil {
		return true
	}

	return *c.EnableCheckAssociateMilestone
}
