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

	// Bitbucket permissions
	BitbucketProjectAdmin = "PROJECT_ADMIN"
	BitbucketProjectWrite = "PROJECT_WRITE"
	BitbucketProjectRead  = "PROJECT_READ"
	BitbucketRepoAdmin    = "REPO_ADMIN"
	BitbucketRepoWrite    = "REPO_WRITE"
	BitbucketRepoRead     = "REPO_READ"
	BitbucketRepoCreate   = "REPO_CREATE"
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
