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
func NewClient(cfg *Config) *Client {
	var err error
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	if cfg.SkipVerify {
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
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
			return nil
		}
	}

	return &Client{
		gh:     ghClient,
		logger: cfg.Logger,
	}
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
