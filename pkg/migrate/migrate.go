package migrate

import (
	"log/slog"

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
}

// CreateNewOrg create new organization
func (m *migrate) CreateNewOrg(opts CreateNewOrgOption) error {
	visibility := gsdk.VisibleTypePrivate
	switch opts.Public {
	case true:
		visibility = gsdk.VisibleTypePublic
	case false:
		visibility = gsdk.VisibleTypePrivate
	}

	m.logger.Info("start create organization", "name", opts.Name)
	_, err := m.gtClient.CreateAndGetOrg(gitea.CreateOrgOption{
		Name:        opts.Name,
		Description: opts.Description,
		Visibility:  visibility,
	})
	if err != nil {
		return err
	}

	m.logger.Info("start migrate organization permission", "name", opts.Name)
	for permission, users := range opts.Permission {
		team, err := m.gtClient.CreateOrGetTeam(opts.Name, permission)
		if err != nil {
			return err
		}
		for _, user := range users {
			err := m.gtClient.AddTeamMember(team.ID, user)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
