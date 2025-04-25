package migrate

import (
	"context"
	"log/slog"
	"regexp"

	"github.com/appleboy/com/convert"
	"github.com/appleboy/github2gitea/pkg/gitea"
	"github.com/appleboy/github2gitea/pkg/github"

	gsdk "code.gitea.io/sdk/gitea"
)

type migrate struct {
	ghClient *github.Client
	gtClient *gitea.Client
	logger   *slog.Logger
}

func New(ghClient *github.Client, gtClient *gitea.Client, logger *slog.Logger) *migrate {
	return &migrate{
		ghClient: ghClient,
		gtClient: gtClient,
		logger:   logger,
	}
}

// CreateNewOrgOption create new organization option
type CreateNewOrgOption struct {
	Name        string
	Description string
	Public      bool
	Permission  map[string][]string
	SourceID    int64
}

// CreateNewOrg create new organization
var invalidCharsRegex = regexp.MustCompile(`[^a-zA-Z0-9\-_\.]`)

func (m *migrate) CreateNewOrg(ctx context.Context, opts CreateNewOrgOption) (*gsdk.Organization, error) {
	visibility := gsdk.VisibleTypePrivate
	if opts.Public {
		visibility = gsdk.VisibleTypePublic
	}

	m.logger.Info("start create organization", "name", opts.Name)
	org, err := m.gtClient.CreateAndGetOrg(gitea.CreateOrgOption{
		Name:        opts.Name,
		Description: opts.Description,
		Visibility:  visibility,
	})
	if err != nil {
		return nil, err
	}

	// get github organization members
	ghUsers, err := m.ghClient.ListOrgUsers(ctx, opts.Name)
	if err != nil {
		return nil, err
	}
	// create gitea organization members
	for _, ghUser := range ghUsers {
		// get github user
		ghUser, err := m.ghClient.GetUser(ctx, convert.FromPtr(ghUser.Login))
		if err != nil {
			m.logger.Error(
				"failed to get github user",
				"name", convert.FromPtr(ghUser.Login),
				"error", err,
			)
			continue
		}

		// create gitea user
		gtUser, err := m.gtClient.CreateOrGetUser(gitea.CreateUserOption{
			LoginName: convert.FromPtr(ghUser.Login),
			Username:  convert.FromPtr(ghUser.Login),
			FullName:  convert.FromPtr(ghUser.Name),
			Email:     convert.FromPtr(ghUser.Email),
			SourceID:  opts.SourceID,
		})
		if err != nil {
			m.logger.Error(
				"failed to create gitea user",
				"name", convert.FromPtr(ghUser.Login),
				"error", err,
			)
			continue
		}

		// get github user permission from org
		_, err = m.ghClient.GetUserPermissionFromOrg(ctx, opts.Name, gtUser.LoginName)
		if err != nil {
			m.logger.Error(
				"failed to get github user permission",
				"name", convert.FromPtr(ghUser.Login),
				"error", err,
			)
			continue
		}
	}

	// get github organization teams
	ghTeams, err := m.ghClient.ListOrgTeams(ctx, opts.Name)
	if err != nil {
		return nil, err
	}
	// create gitea organization teams
	for _, ghTeam := range ghTeams {
		// Sanitize the team name
		sanitizedTeamName := invalidCharsRegex.ReplaceAllString(convert.FromPtr(ghTeam.Name), "_")
		team, err := m.gtClient.CreateOrGetTeam(opts.Name, gitea.CreateTeamOption{
			Name:        sanitizedTeamName,
			Description: convert.FromPtr(ghTeam.Description),
			Permission:  convert.FromPtr(ghTeam.Permission),
		})
		if err != nil {
			m.logger.Error(
				"failed to create gitea team",
				"name", convert.FromPtr(ghTeam.Name),
				"error", err,
			)
			continue
		}

		// get github team members
		ghUsers, err := m.ghClient.ListOrgTeamsMembers(ctx, opts.Name, *ghTeam.Slug)
		if err != nil {
			m.logger.Error(
				"failed to get github team members",
				"name", convert.FromPtr(ghTeam.Name),
				"error", err,
			)
			continue
		}

		// add gitea team members
		for _, ghUser := range ghUsers {
			err := m.gtClient.AddTeamMember(team.ID, convert.FromPtr(ghUser.Login))
			if err != nil {
				m.logger.Error(
					"failed to add gitea team member",
					"name", convert.FromPtr(ghTeam.Name),
					"user", convert.FromPtr(ghUser.Login),
					"error", err,
				)
				continue
			}
		}
	}
	return org, nil
}

// MigrateNewRepoOption migrate repository option
type MigrateNewRepoOption struct {
	Owner        string
	Name         string
	CloneAddr    string
	Description  string
	Private      bool
	Permission   map[string][]string
	AuthUsername string
	AuthToken    string
}

// MigrateNewRepo migrate repository
func (m *migrate) MigrateNewRepo(ctx context.Context, opts MigrateNewRepoOption) error {
	m.logger.Info("start migrate repo",
		"owner", opts.Owner,
		"name", opts.Name,
	)
	_, err := m.gtClient.MigrateRepo(gitea.MigrateRepoOption{
		RepoName:     opts.Name,
		RepoOwner:    opts.Owner,
		CloneAddr:    opts.CloneAddr,
		Private:      opts.Private,
		Description:  opts.Description,
		AuthUsername: opts.AuthUsername,
		AuthToken:    opts.AuthToken,
	})
	if err != nil {
		return err
	}

	// List collaborators
	ghUsers, err := m.ghClient.ListRepoCollaborators(ctx, opts.Owner, opts.Name)
	if err != nil {
		return err
	}

	for _, ghUser := range ghUsers {
		if *ghUser.Type != "User" {
			m.logger.Info(
				"skip github user type",
				"name", convert.FromPtr(ghUser.Login),
				"type", convert.FromPtr(ghUser.Type),
			)
			continue
		}

		// get github user
		ghUser, err := m.ghClient.GetUser(ctx, convert.FromPtr(ghUser.Login))
		if err != nil {
			m.logger.Error(
				"failed to get github user",
				"name", convert.FromPtr(ghUser.Login),
				"error", err,
			)
			continue
		}

		// add gitea collaborator
		_, err = m.gtClient.AddCollaborator(opts.Owner, opts.Name, *ghUser.Login, ghUser.Permissions)
		if err != nil {
			m.logger.Error(
				"failed to add gitea repo collaborator",
				"name", convert.FromPtr(ghUser.Login),
				"error", err,
			)
			continue
		}
	}

	m.logger.Info("migrate repo success",
		"owner", opts.Owner,
		"name", opts.Name,
	)

	return nil
}
