package github

import (
	"context"
	"crypto/tls"
	"errors"
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
	if cfg == nil {
		return nil, errors.New("github config is required")
	}
	if cfg.Token == "" {
		return nil, errors.New("github token is required")
	}
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

/*
ListRepoUsers lists all users with access to a repository.
This is now implemented using paginatedFetch.
*/
func (c *Client) ListRepoUsers(ctx context.Context, owner, repo string) ([]*github.User, error) {
	return paginatedFetch(ctx, func(page int) ([]*github.User, *github.Response, error) {
		return c.gh.Repositories.ListCollaborators(ctx, owner, repo, &github.ListCollaboratorsOptions{
			ListOptions: github.ListOptions{
				Page:    page,
				PerPage: 100,
			},
		})
	})
}

/*
ListRepoCollaborators lists all collaborators in a repository.
This is now implemented using paginatedFetch.
*/
func (c *Client) ListRepoCollaborators(ctx context.Context, owner, repo string) ([]*github.User, error) {
	return paginatedFetch(ctx, func(page int) ([]*github.User, *github.Response, error) {
		return c.gh.Repositories.ListCollaborators(ctx, owner, repo, &github.ListCollaboratorsOptions{
			ListOptions: github.ListOptions{
				Page:    page,
				PerPage: 100,
			},
		})
	})
}

/*
paginatedFetch is a generic helper for paginated GitHub API calls.
fetch: a function that takes a page number and returns items, response, error.
*/
func paginatedFetch[T any](
	_ context.Context,
	fetch func(page int) ([]*T, *github.Response, error),
) ([]*T, error) {
	var allItems []*T
	page := 1
	for {
		items, resp, err := fetch(page)
		if err != nil {
			return nil, err
		}
		allItems = append(allItems, items...)
		if resp == nil || resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}
	return allItems, nil
}

// ListOrgTeams lists all teams in an organization
// permission can be one of: "pull", "triage", "push", "maintain", "admin"
func (c *Client) ListOrgTeams(ctx context.Context, org string) ([]*github.Team, error) {
	return paginatedFetch(ctx, func(page int) ([]*github.Team, *github.Response, error) {
		return c.gh.Teams.ListTeams(ctx, org, &github.ListOptions{
			Page:    page,
			PerPage: 100,
		})
	})
}

// ListOrgTeamsMembers lists all members in a team using paginatedFetch
func (c *Client) ListOrgTeamsMembers(ctx context.Context, org string, slug string) ([]*github.User, error) {
	return paginatedFetch(ctx, func(page int) ([]*github.User, *github.Response, error) {
		return c.gh.Teams.ListTeamMembersBySlug(ctx, org, slug, &github.TeamListTeamMembersOptions{
			ListOptions: github.ListOptions{
				Page:    page,
				PerPage: 100,
			},
		})
	})
}

// ListOrgUsers lists all members in an organization using paginatedFetch
func (c *Client) ListOrgUsers(ctx context.Context, org string) ([]*github.User, error) {
	return paginatedFetch(ctx, func(page int) ([]*github.User, *github.Response, error) {
		return c.gh.Organizations.ListMembers(ctx, org, &github.ListMembersOptions{
			ListOptions: github.ListOptions{
				Page:    page,
				PerPage: 100,
			},
		})
	})
}

// ListOrgRepos lists all repositories in an organization using paginatedFetch
func (c *Client) ListOrgRepos(ctx context.Context, org string) ([]*github.Repository, error) {
	return paginatedFetch(ctx, func(page int) ([]*github.Repository, *github.Response, error) {
		return c.gh.Repositories.ListByOrg(ctx, org, &github.RepositoryListByOrgOptions{
			ListOptions: github.ListOptions{
				Page:    page,
				PerPage: 10,
			},
		})
	})
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
