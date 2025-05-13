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
	OldName     string
	NewName     string
	Description string
	Public      bool
	Permission  map[string][]string
	SourceID    int64
}

// CreateNewOrgResult create new organization result
type CreateNewOrgResult struct {
	Org       *gsdk.Organization
	Admins    []*gsdk.User
	RepoTeams map[string][]*gsdk.Team
}

// CreateNewOrg create new organization
var invalidCharsRegex = regexp.MustCompile(`[^a-zA-Z0-9\-_\.]`)

func (m *migrate) CreateNewOrg(ctx context.Context, opts CreateNewOrgOption) (*CreateNewOrgResult, error) {
	visibility := gsdk.VisibleTypePrivate
	if opts.Public {
		visibility = gsdk.VisibleTypePublic
	}

	m.logger.Info("start create organization", "name", opts.NewName)
	org, err := m.gtClient.CreateAndGetOrg(gitea.CreateOrgOption{
		Name:        opts.NewName,
		Description: opts.Description,
		Visibility:  visibility,
	})
	if err != nil {
		return nil, err
	}

	owners, err := m.gtClient.SearchOrgTeams(org.UserName, &gsdk.SearchTeamsOptions{
		Query: "owners",
	})
	if err != nil {
		return nil, err
	}
	ownerTeam := owners[0]

	// get github organization members
	ghUsers, err := m.ghClient.ListOrgUsers(ctx, opts.OldName)
	if err != nil {
		return nil, err
	}

	admins := make([]*gsdk.User, 0)
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

		// Role identifies the user's role within the organization or team.
		// Possible values for organization membership:
		//     member - non-owner organization member
		//     admin - organization owner
		//
		// Possible values for team membership are:
		//     member - a normal member of the team
		//     maintainer - a team maintainer. Able to add/remove other team
		//                  members, promote other team members to team
		//                  maintainer, and edit the team’s name and description
		role, err := m.ghClient.GetUserPermissionFromOrg(ctx, opts.OldName, gtUser.LoginName)
		if err != nil {
			m.logger.Error(
				"failed to get github user permission",
				"name", convert.FromPtr(ghUser.Login),
				"error", err,
			)
			continue
		}

		if role == "admin" {
			admins = append(admins, gtUser)
			err := m.gtClient.AddTeamMember(ownerTeam.ID, gtUser.UserName)
			if err != nil {
				m.logger.Error(
					"failed to add gitea team member (admin)",
					"name", ownerTeam.Name,
					"user", gtUser.UserName,
					"error", err,
				)
				continue
			}
		}
	}

	repoTeams := make(map[string][]*gsdk.Team)
	// get github organization teams
	ghTeams, err := m.ghClient.ListOrgTeams(ctx, opts.OldName)
	if err != nil {
		return nil, err
	}
	// create gitea organization teams
	for _, ghTeam := range ghTeams {
		// get github team repositories
		ghRepos, err := m.ghClient.ListTeamReposBySlug(ctx, opts.OldName, *ghTeam.Slug)
		if err != nil {
			m.logger.Error(
				"failed to get github team repositories",
				"name", convert.FromPtr(ghTeam.Name),
				"error", err,
			)
			continue
		}

		// Sanitize the team name
		sanitizedTeamName := invalidCharsRegex.ReplaceAllString(convert.FromPtr(ghTeam.Name), "_")
		team, err := m.gtClient.CreateOrGetTeam(opts.NewName, gitea.CreateTeamOption{
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

		for _, ghRepo := range ghRepos {
			repoTeams[convert.FromPtr(ghRepo.Name)] = append(repoTeams[convert.FromPtr(ghRepo.Name)], team)
		}

		m.logger.Info("create gitea team",
			"org", opts.NewName,
			"name", team.Name,
			"permission", team.Permission,
		)

		// get github team members
		ghUsers, err := m.ghClient.ListOrgTeamsMembers(ctx, opts.OldName, *ghTeam.Slug)
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

	resp := &CreateNewOrgResult{
		Org:       org,
		Admins:    admins,
		RepoTeams: repoTeams,
	}

	return resp, nil
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

	m.logger.Info("migrate repo success",
		"owner", opts.Owner,
		"name", opts.Name,
	)

	return nil
}
