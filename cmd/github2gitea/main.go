package main

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/appleboy/github2gitea/pkg/config"
	gt "github.com/appleboy/github2gitea/pkg/gitea"
	gh "github.com/appleboy/github2gitea/pkg/github"
	"github.com/appleboy/github2gitea/pkg/migrate"

	gsdk "code.gitea.io/sdk/gitea"
	"github.com/appleboy/com/convert"
	"github.com/google/go-github/v71/github"
)

func setupLogger(debug bool) *slog.Logger {
	logLevel := slog.LevelInfo
	if debug {
		logLevel = slog.LevelDebug
	}
	return slog.New(slog.NewTextHandler(log.Writer(), &slog.HandlerOptions{
		Level: logLevel,
	}))
}

func createClients(ctx context.Context, cfg *config.Config, logger *slog.Logger) (ghClient *gh.Client, gtClient *gt.Client, err error) {
	ghClient, err = gh.NewClient(&gh.Config{
		Token:      cfg.GHToken,
		Server:     cfg.GHServer,
		SkipVerify: cfg.GHSkipVerify,
		Logger:     logger,
	})
	if err != nil {
		return nil, nil, err
	}

	gtClient, err = gt.New(ctx, &gt.Config{
		Server:     cfg.GTServer,
		Token:      cfg.GTToken,
		SkipVerify: cfg.GTSkipVerify,
		Logger:     logger,
		SourceID:   cfg.GTSourceID,
	})
	if err != nil {
		return nil, nil, err
	}
	return ghClient, gtClient, nil
}

func printUserInfo(logger *slog.Logger, ghUser *github.User, gtUser *gsdk.User) {
	logger.Info("github user",
		"login", convert.FromPtr(ghUser.Login),
		"name", convert.FromPtr(ghUser.Name),
		"email", convert.FromPtr(ghUser.Email),
	)
	logger.Info("gitea user",
		"login", gtUser.UserName,
		"name", gtUser.FullName,
		"email", gtUser.Email,
	)
}

func migrateOrgAndRepos(ctx context.Context, cfg *config.Config, logger *slog.Logger, ghClient *gh.Client, gtClient *gt.Client) error {
	// get github current user
	ghUser, err := ghClient.GetCurrentUser(ctx)
	if err != nil {
		logger.Error("failed to get current github user", "error", err)
		return err
	}

	// get gitea current user
	gtUser, err := gtClient.GetCurrentUser()
	if err != nil {
		logger.Error("failed to get current gitea user", "error", err)
		return err
	}

	printUserInfo(logger, ghUser, gtUser)

	// get github organization
	ghOrg, err := ghClient.GetOrg(ctx, cfg.SourceOrg)
	if err != nil {
		logger.Error("failed to get github org", "error", err)
		return err
	}

	m := migrate.New(
		ghClient,
		gtClient,
		logger,
	)

	// create new gitea organization
	_, err = m.CreateNewOrg(ctx, migrate.CreateNewOrgOption{
		Name:        cfg.TargetOrg,
		Description: convert.FromPtr(ghOrg.Description),
		Public:      false,
		SourceID:    cfg.GTSourceID,
	})
	if err != nil {
		logger.Error("failed to create gitea org", "error", err)
		return err
	}

	// get github repo list from organization
	ghRepos, err := ghClient.ListOrgRepos(ctx, cfg.SourceOrg)
	if err != nil {
		logger.Error("failed to get github org repos", "error", err)
		return err
	}

	for _, repo := range ghRepos {
		// create new gitea repository
		err = m.MigrateNewRepo(ctx, migrate.MigrateNewRepoOption{
			Owner:        convert.FromPtr(repo.Owner.Login),
			Name:         convert.FromPtr(repo.Name),
			CloneAddr:    convert.FromPtr(repo.CloneURL),
			Description:  convert.FromPtr(repo.Description),
			Private:      convert.FromPtr(repo.Private),
			AuthUsername: convert.FromPtr(ghUser.Login),
			AuthToken:    cfg.GHToken,
		})
		if err != nil {
			logger.Error("migration repository error", "error", err)
		}
	}
	return nil
}

func main() {
	cfg := config.LoadConfig()
	logger := setupLogger(cfg.Debug)

	if cfg.SourceOrg == "" || cfg.TargetOrg == "" {
		logger.Error("source or target org is empty")
		return
	}

	// check timeout format
	timeout, err := time.ParseDuration(cfg.APITimeout)
	if err != nil {
		logger.Error("failed to parse timeout", "error", err)
		return
	}
	// command timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ghClient, gtClient, err := createClients(ctx, cfg, logger)
	if err != nil {
		logger.Error("failed to create clients", "error", err)
		return
	}

	if err := migrateOrgAndRepos(ctx, cfg, logger, ghClient, gtClient); err != nil {
		logger.Error("migration failed", "error", err)
	}
}
