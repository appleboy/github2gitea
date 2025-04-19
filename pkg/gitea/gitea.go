package gitea

import (
	"context"
	"crypto/tls"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/appleboy/github2gitea/pkg/core"

	gsdk "code.gitea.io/sdk/gitea"
)

type Config struct {
	Server     string
	Token      string
	SkipVerify bool
	SourceID   int64
	Logger     *slog.Logger
}

// NewGitea creates a new instance of the gitea struct.
func NewGitea(ctx context.Context, cfg *Config) (*gitea, error) {
	g := &gitea{
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

// gitea is a struct that holds the gitea client.
type gitea struct {
	ctx        context.Context
	server     string
	token      string
	skipVerify bool
	sourceID   int64
	client     *gsdk.Client
	logger     *slog.Logger
}

// init initializes the gitea client.
func (g *gitea) init() error {
	if g.server == "" || g.token == "" {
		return errors.New("missing gitea server or token")
	}

	g.server = strings.TrimRight(g.server, "/")

	opts := []gsdk.ClientOption{
		gsdk.SetToken(g.token),
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

// CreateOrgOption create organization option
type CreateOrgOption struct {
	Name        string
	Description string
	Visibility  gsdk.VisibleType
}

// CreateAndGetOrg creates or retrieves an organization, handling error cases properly
func (g *gitea) CreateAndGetOrg(opts CreateOrgOption) (*gsdk.Organization, error) {
	newOrg, response, err := g.client.GetOrg(opts.Name)
	if err != nil {
		// Handle 404 case by creating the organization
		if response != nil && response.StatusCode == http.StatusNotFound {
			visible := opts.Visibility
			newOrg, _, err = g.client.CreateOrg(gsdk.CreateOrgOption{
				Name:        opts.Name,
				Description: opts.Description,
				Visibility:  visible,
			})
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return newOrg, nil
}

// MigrateRepoOption migrate repository option
type MigrateRepoOption struct {
	RepoName     string
	RepoOwner    string
	CloneAddr    string
	Private      bool
	Description  string
	AuthUsername string
	AuthPassword string
}

// MigrateRepo migrate repository
func (g *gitea) MigrateRepo(opts MigrateRepoOption) (*gsdk.Repository, error) {
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
		AuthPassword: opts.AuthPassword,
	})
	if err != nil {
		return nil, err
	}

	return newRepo, nil
}

type CreateUserOption struct {
	SourceID  int64
	LoginName string
	Username  string
	FullName  string
	Email     string
}

// CreateOrGetUser create or get user
func (g *gitea) CreateOrGetUser(opts CreateUserOption) (*gsdk.User, error) {
	user, resp, err := g.client.GetUserInfo(opts.Username)
	if err != nil {
		if g.logger != nil {
			g.logger.Warn("get user info failed", "username", opts.Username, "err", err)
		}
	}
	if resp.StatusCode == http.StatusNotFound {
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
			return nil, err
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

// AddCollaborator add collaborator
func (g *gitea) AddCollaborator(org, repo, user, permission string) (*gsdk.Response, error) {
	var access gsdk.AccessMode
	switch permission {
	case core.GiteaRepoAdmin:
		access = gsdk.AccessModeAdmin
	case core.GiteaRepoWrite:
		access = gsdk.AccessModeWrite
	case core.GiteaRepoRead:
		access = gsdk.AccessModeRead
	default:
		return nil, errors.New("permission mode invalid")
	}
	return g.client.AddCollaborator(org, repo, user, gsdk.AddCollaboratorOption{
		Permission: &access,
	})
}

// CreateOrGetTeam create team
func (g *gitea) CreateOrGetTeam(org, permission string) (*gsdk.Team, error) {
	var opt gsdk.CreateTeamOption
	switch permission {
	case core.GiteaProjectAdmin:
		opt = gsdk.CreateTeamOption{
			Name:                    "OrgAdmin",
			Description:             "OrgAdmin",
			Permission:              gsdk.AccessModeAdmin,
			IncludesAllRepositories: true,
			CanCreateOrgRepo:        true,
			Units:                   core.DefaultUnits,
		}
	case core.GiteaProjectWrite:
		opt = gsdk.CreateTeamOption{
			Name:                    "OrgWriter",
			Description:             "OrgWriter",
			Permission:              gsdk.AccessModeWrite,
			IncludesAllRepositories: true,
			Units:                   core.DefaultUnits,
		}
	case core.GiteaProjectRead:
		opt = gsdk.CreateTeamOption{
			Name:                    "OrgReader",
			Description:             "OrgReader",
			Permission:              gsdk.AccessModeRead,
			IncludesAllRepositories: true,
			Units:                   core.DefaultUnits,
		}
	case core.GiteaRepoCreate:
		opt = gsdk.CreateTeamOption{
			Name:                    "RepoCreator",
			Description:             "RepoCreator",
			Permission:              gsdk.AccessModeRead,
			IncludesAllRepositories: false,
			CanCreateOrgRepo:        true,
			Units:                   core.DefaultUnits,
		}
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

// AddTeamMember add team member
func (g *gitea) AddTeamMember(id int64, user string) error {
	_, err := g.client.AddTeamMember(id, user)
	return err
}
