package core

import gsdk "code.gitea.io/sdk/gitea"

const (
	// Gitea permissions
	GiteaRepoAdmin    = "admin"
	GiteaRepoWrite    = "write"
	GiteaRepoRead     = "read"
	GiteaProjectAdmin = "admin"
	GiteaProjectWrite = "write"
	GiteaProjectRead  = "read"
	GiteaRepoCreate   = "create"

	GitHubTeamPull     = "pull"
	GitHubTeamPush     = "push"
	GitHubTeamAdmin    = "admin"
	GitHubTeamMaintain = "maintain"
	GitHubTeamTriager  = "triager"
)

var DefaultUnits = []gsdk.RepoUnitType{
	gsdk.RepoUnitCode,
	gsdk.RepoUnitIssues,
	gsdk.RepoUnitExtIssues,
	gsdk.RepoUnitExtWiki,
	gsdk.RepoUnitPackages,
	gsdk.RepoUnitProjects,
	gsdk.RepoUnitPulls,
	gsdk.RepoUnitReleases,
	gsdk.RepoUnitWiki,
	gsdk.RepoUnitActions,
}
