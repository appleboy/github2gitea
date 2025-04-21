package gitea

// Package gitea provides a client and helper functions for interacting with a Gitea server.

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/appleboy/github2gitea/pkg/core"

	gsdk "code.gitea.io/sdk/gitea"
)

// GiteaError is a custom error type for Gitea operations.
type GiteaError struct {
	// Operation is the name of the Gitea operation that failed.
	Operation string
	// Code is the HTTP status code returned by the Gitea server.
	Code int
	// Message is the error message returned by the Gitea server.
	Message string
}

func (e *GiteaError) Error() string {
	// Error implements the error interface for GiteaError.
	return fmt.Sprintf("gitea %s failed: [%d] %s", e.Operation, e.Code, e.Message)
}

type Config struct {
	// Server is the Gitea server URL.
	Server string
	// Token is the personal access token for authentication.
	Token string
	// SkipVerify determines whether to skip TLS certificate verification.
	SkipVerify bool
	// SourceID is the authentication source ID for user creation.
	SourceID int64
	// Logger is the logger instance for logging.
	Logger *slog.Logger
}

// New creates a new Gitea client with the provided configuration and context.
// Returns a pointer to the Client and an error if initialization fails.
func New(ctx context.Context, cfg *Config) (*Client, error) {
	if cfg.Server == "" || (!strings.HasPrefix(cfg.Server, "http://") && !strings.HasPrefix(cfg.Server, "https://")) {
		return nil, errors.New("invalid gitea server: must start with http:// or https://")
	}

	g := &Client{
		ctx:        ctx,
		server:     cfg.Server,
		token:      cfg.Token,
		skipVerify: cfg.SkipVerify,
		sourceID:   cfg.SourceID,
		logger:     cfg.Logger,
	}

	err := g.init()
	if err != nil {
		return nil, err
	}

	return g, nil
}

// Client represents a Gitea client instance for interacting with the Gitea API.
type Client struct {
	ctx        context.Context
	server     string
	token      string
	skipVerify bool
	sourceID   int64
	client     *gsdk.Client
	logger     *slog.Logger
}

// init initializes the underlying Gitea SDK client.
// Returns an error if the server or token is missing, or if client creation fails.
func (g *Client) init() error {
	if g.server == "" || g.token == "" {
		return errors.New("missing gitea server or token")
	}

	g.server = strings.TrimRight(g.server, "/")

	opts := []gsdk.ClientOption{
		gsdk.SetToken(g.token),
		gsdk.SetContext(g.ctx),
		gsdk.SetUserAgent("github2gitea"),
	}

	if g.skipVerify {
		// add new http client for skip verify
		httpClient := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
			},
		}
		opts = append(opts, gsdk.SetHTTPClient(httpClient))
	}

	client, err := gsdk.NewClient(g.server, opts...)
	if err != nil {
		return err
	}
	g.client = client

	return nil
}

// GetCurrentUser retrieves the current authenticated user's information from Gitea.
// Returns a pointer to the User and an error if the request fails.
func (g *Client) GetCurrentUser() (*gsdk.User, error) {
	user, _, err := g.client.GetMyUserInfo()
	if err != nil {
		return nil, err
	}
	return user, nil
}

// CreateOrgOption contains options for creating a Gitea organization.
type CreateOrgOption struct {
	// Name is the organization name.
	Name string
	// Description is the organization description.
	Description string
	// Visibility sets the visibility of the organization.
	Visibility gsdk.VisibleType
}

// CreateAndGetOrg retrieves an existing organization or creates a new one if it does not exist.
// Returns a pointer to the Organization and an error if the operation fails.
func (g *Client) CreateAndGetOrg(opts CreateOrgOption) (*gsdk.Organization, error) {
	newOrg, response, err := g.client.GetOrg(opts.Name)
	if err != nil {
		switch {
		case response != nil && response.StatusCode == http.StatusNotFound:
			// Handle 404 case by creating the organization
			visible := opts.Visibility
			var createErr error
			newOrg, _, createErr = g.client.CreateOrg(gsdk.CreateOrgOption{
				Name:        opts.Name,
				Description: opts.Description,
				Visibility:  visible,
			})
			if createErr != nil {
				// Use the original 404 status code as per the original logic
				return nil, &GiteaError{Operation: "create_org", Code: response.StatusCode, Message: createErr.Error()}
			}
			// If creation succeeded, reset err so we return the new org below
			err = nil
		case response != nil:
			// Handle other errors from GetOrg that have a response
			return nil, &GiteaError{Operation: "get_org", Code: response.StatusCode, Message: err.Error()}
		default: // response == nil
			// Handle errors from GetOrg without a response
			return nil, err
		}
		// If err was non-nil initially but creation succeeded, err is now nil.
		// If any return occurred within the switch, we won't reach here.
	}

	return newOrg, nil
}

// MigrateRepoOption contains options for migrating a repository to Gitea.
type MigrateRepoOption struct {
	// RepoName is the name of the repository to create.
	RepoName string
	// RepoOwner is the owner (user or org) of the new repository.
	RepoOwner string
	// CloneAddr is the source repository clone URL.
	CloneAddr string
	// Private determines if the new repository is private.
	Private bool
	// Description is the repository description.
	Description string
	// AuthUsername is the username for authentication to the source repository.
	AuthUsername string
	// AuthToken is the token/password for authentication to the source repository.
	AuthToken string
}

// MigrateRepo migrates a repository from a remote source to Gitea.
// Returns a pointer to the new Repository and an error if the migration fails.
func (g *Client) MigrateRepo(opts MigrateRepoOption) (*gsdk.Repository, error) {
	if opts.RepoName == "" || opts.RepoOwner == "" || opts.CloneAddr == "" {
		return nil, errors.New("missing required migration parameters: RepoName, RepoOwner and CloneAddr are required")
	}
	newRepo, _, err := g.client.MigrateRepo(gsdk.MigrateRepoOption{
		RepoName:     opts.RepoName,
		RepoOwner:    opts.RepoOwner,
		CloneAddr:    opts.CloneAddr,
		Private:      opts.Private,
		Description:  opts.Description,
		AuthUsername: opts.AuthUsername,
		AuthToken:    opts.AuthToken,
		Service:      gsdk.GitServiceGithub,
		Wiki:         true,
		Milestones:   true,
		Issues:       true,
		Releases:     true,
		Labels:       true,
		PullRequests: true,
	})
	if err != nil {
		return nil, err
	}

	return newRepo, nil
}

// CreateUserOption contains options for creating a Gitea user.
type CreateUserOption struct {
	// SourceID is the authentication source ID.
	SourceID int64
	// LoginName is the login name for the user.
	LoginName string
	// Username is the username for the user.
	Username string
	// FullName is the full name of the user.
	FullName string
	// Email is the email address of the user.
	Email string
}

// CreateOrGetUser retrieves an existing user or creates a new one if not found.
// Returns a pointer to the User and an error if the operation fails.
func (g *Client) CreateOrGetUser(opts CreateUserOption) (*gsdk.User, error) {
	user, resp, err := g.client.GetUserInfo(opts.Username)
	if err != nil {
		if g.logger != nil {
			g.logger.Warn("get user info failed", "username", opts.Username, "err", err)
		}
		if resp != nil && resp.StatusCode != http.StatusNotFound {
			return nil, &GiteaError{Operation: "get_user_info", Code: resp.StatusCode, Message: err.Error()}
		}
	}
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		mustChangePassword := false
		user, _, err = g.client.AdminCreateUser(gsdk.CreateUserOption{
			SourceID:           opts.SourceID,
			LoginName:          opts.LoginName,
			Username:           opts.Username,
			FullName:           opts.FullName,
			Email:              opts.Email,
			MustChangePassword: &mustChangePassword,
		})
		if err != nil {
			return nil, &GiteaError{Operation: "admin_create_user", Code: http.StatusInternalServerError, Message: err.Error()}
		}
		if g.logger != nil {
			g.logger.Info(
				"create a new user",
				"username", opts.Username,
				"fullname", opts.FullName,
			)
		}
	}

	return user, nil
}

// AddCollaborator adds a user as a collaborator to the specified repository with the given permissions.
// Returns the response and an error if the operation fails.
func (g *Client) AddCollaborator(org, repo, user string, permission map[string]bool) (*gsdk.Response, error) {
	var access gsdk.AccessMode
	switch {
	case permission[core.GitHubTeamAdmin]:
		access = gsdk.AccessModeAdmin
	case permission[core.GitHubTeamMaintain]:
		access = gsdk.AccessModeOwner
	case permission[core.GitHubTeamPush]:
		access = gsdk.AccessModeWrite
	case permission[core.GitHubTeamPull]:
		access = gsdk.AccessModeRead
	default:
		// Default to read access if no specific permission is set or recognized
		access = gsdk.AccessModeRead
	}
	return g.client.AddCollaborator(org, repo, user, gsdk.AddCollaboratorOption{
		Permission: &access,
	})
}

// CreateTeamOption contains options for creating a Gitea team.
type CreateTeamOption struct {
	// Name is the team name.
	Name string
	// Description is the team description.
	Description string
	// Permission is the permission level for the team.
	Permission string
}

// CreateOrGetTeam retrieves an existing team or creates a new one in the specified organization.
// Returns a pointer to the Team and an error if the operation fails.
func (g *Client) CreateOrGetTeam(org string, opts CreateTeamOption) (*gsdk.Team, error) {
	opt := gsdk.CreateTeamOption{
		Name:        opts.Name,
		Description: opts.Description,
		Permission:  gsdk.AccessMode(opts.Permission),
		Units:       core.DefaultUnits,
	}

	switch opts.Permission {
	case core.GitHubTeamAdmin:
		opt.Permission = gsdk.AccessModeAdmin
		opt.CanCreateOrgRepo = true
	case core.GitHubTeamPush:
		opt.Permission = gsdk.AccessModeWrite
	case core.GitHubTeamPull:
		opt.Permission = gsdk.AccessModeRead
	case core.GitHubTeamMaintain:
		opt.Permission = gsdk.AccessModeWrite
	case core.GitHubTeamTriager: // not supported
	default:
		return nil, errors.New("permission mode invalid")
	}

	teams, _, err := g.client.SearchOrgTeams(org, &gsdk.SearchTeamsOptions{
		Query: opt.Name,
	})
	if err != nil {
		return nil, err
	}
	if len(teams) > 0 {
		return teams[0], nil
	}

	// create team
	team, _, err := g.client.CreateTeam(org, opt)
	if err != nil {
		return nil, err
	}

	return team, nil
}

// AddTeamMember adds a user to the specified team by team ID.
// Returns an error if the operation fails.
func (g *Client) AddTeamMember(id int64, user string) error {
	_, err := g.client.AddTeamMember(id, user)
	return err
}
