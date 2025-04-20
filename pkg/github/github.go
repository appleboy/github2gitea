package github

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/go-github/v71/github"
)

type Config struct {
	Server     string
	Token      string
	SkipVerify bool
	Logger     *slog.Logger
}

// Client wraps the GitHub client with additional methods
type Client struct {
	logger *slog.Logger
	gh     *github.Client
}

// NewClient creates a new GitHub Client
func NewClient(cfg *Config) (*Client, error) {
	var err error
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	if cfg.SkipVerify {
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
		}
	}

	ghClient := github.NewClient(httpClient).
		WithAuthToken(cfg.Token)

	if cfg.Server != "" {
		ghClient, err = ghClient.WithEnterpriseURLs(
			cfg.Server,
			cfg.Server,
		)
		if err != nil {
			cfg.Logger.Error("failed to create GitHub client", "error", err)
			return nil, err
		}
	}

	return &Client{
		gh:     ghClient,
		logger: cfg.Logger,
	}, nil
}

// GetUser gets a user's information by username
func (c *Client) GetUser(ctx context.Context, username string) (*github.User, error) {
	user, _, err := c.gh.Users.Get(ctx, username)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetCurrentUser gets the current authenticated user's information
func (c *Client) GetCurrentUser(ctx context.Context) (*github.User, error) {
	user, _, err := c.gh.Users.Get(ctx, "")
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetUserPermissionFromRepo gets a user's permission level for a repository
func (c *Client) GetUserPermissionFromRepo(ctx context.Context, owner, repo, username string) (string, error) {
	permission, _, err := c.gh.Repositories.GetPermissionLevel(ctx, owner, repo, username)
	if err != nil {
		return "", err
	}
	return permission.GetPermission(), nil
}

// GetUserPermissionFromOrg gets a user's permission level in an organization
func (c *Client) GetUserPermissionFromOrg(ctx context.Context, org, username string) (string, error) {
	membership, _, err := c.gh.Organizations.GetOrgMembership(ctx, username, org)
	if err != nil {
		return "", err
	}
	return membership.GetRole(), nil
}

// ListRepoUsers lists all users with access to a repository
func (c *Client) ListRepoUsers(ctx context.Context, owner, repo string) ([]*github.User, error) {
	opts := &github.ListCollaboratorsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allUsers []*github.User
	for {
		users, resp, err := c.gh.Repositories.ListCollaborators(ctx, owner, repo, opts)
		if err != nil {
			return nil, err
		}
		allUsers = append(allUsers, users...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return allUsers, nil
}

// ListRepoCollaborators lists all collaborators in a repository
func (c *Client) ListRepoCollaborators(ctx context.Context, owner, repo string) ([]*github.User, error) {
	opts := &github.ListCollaboratorsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	var allUsers []*github.User
	for {
		users, resp, err := c.gh.Repositories.ListCollaborators(ctx, owner, repo, opts)
		if err != nil {
			return nil, err
		}
		allUsers = append(allUsers, users...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return allUsers, nil
}

// ListOrgTeams lists all teams in an organization
// permission can be one of: "pull", "triage", "push", "maintain", "admin"
func (c *Client) ListOrgTeams(ctx context.Context, org string) ([]*github.Team, error) {
	opts := &github.ListOptions{PerPage: 100}
	var allTeams []*github.Team
	for {
		teams, resp, err := c.gh.Teams.ListTeams(ctx, org, opts)
		if err != nil {
			return nil, err
		}
		allTeams = append(allTeams, teams...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return allTeams, nil
}

// ListOrgTeamsMembers lists all members in a team
func (c *Client) ListOrgTeamsMembers(ctx context.Context, org string, slug string) ([]*github.User, error) {
	opts := &github.TeamListTeamMembersOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	var allMembers []*github.User
	for {
		members, resp, err := c.gh.Teams.ListTeamMembersBySlug(ctx, org, slug, opts)
		if err != nil {
			return nil, err
		}
		allMembers = append(allMembers, members...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return allMembers, nil
}

// ListOrgUsers lists all members in an organization
func (c *Client) ListOrgUsers(ctx context.Context, org string) ([]*github.User, error) {
	opts := &github.ListMembersOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allUsers []*github.User
	for {
		users, resp, err := c.gh.Organizations.ListMembers(ctx, org, opts)
		if err != nil {
			return nil, err
		}
		allUsers = append(allUsers, users...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return allUsers, nil
}

// ListOrgRepos lists all repositories in an organization
func (c *Client) ListOrgRepos(ctx context.Context, org string) ([]*github.Repository, error) {
	opts := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allRepos []*github.Repository
	for {
		repos, resp, err := c.gh.Repositories.ListByOrg(ctx, org, opts)
		if err != nil {
			return nil, err
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return allRepos, nil
}

// GetRepo gets a single repository's information
func (c *Client) GetRepo(ctx context.Context, owner, repo string) (*github.Repository, error) {
	repository, _, err := c.gh.Repositories.Get(ctx, owner, repo)
	return repository, err
}

// GetOrg gets a single organization's information
func (c *Client) GetOrg(ctx context.Context, org string) (*github.Organization, error) {
	organization, _, err := c.gh.Organizations.Get(ctx, org)
	return organization, err
}
