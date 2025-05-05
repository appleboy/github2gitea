package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/appleboy/github2gitea/pkg/config"
	gt "github.com/appleboy/github2gitea/pkg/gitea"
	gh "github.com/appleboy/github2gitea/pkg/github"
	"github.com/appleboy/github2gitea/pkg/migrate"
	"github.com/appleboy/github2gitea/pkg/version"

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
	ghRepos, err := ghClient.ListOrgRepos(ctx, *ghOrg.Login)
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

type UserCSV struct {
	Login string
	Email string
	Role  string
}

func readUserList(path string) ([]UserCSV, error) {
	if path == "" {
		return nil, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	var users []UserCSV
	for index, rec := range records {
		// Skip header and invalid lines
		if index == 0 || len(rec) < 5 {
			continue
		}
		users = append(users, UserCSV{
			Login: rec[2],
			Email: rec[3],
			Role:  rec[4],
		})
	}
	return users, nil
}

func createUsersFromCSV(ctx context.Context, ghClient *gh.Client, gtClient *gt.Client, users []UserCSV, sourceID int64, logger *slog.Logger) {
	for _, u := range users {
		ghUser, err := ghClient.GetUser(ctx, u.Login)
		if err != nil {
			logger.Error("failed to get github user", "login", u.Login, "error", err)
			continue
		}

		opt := gt.CreateUserOption{
			SourceID:  sourceID,
			LoginName: u.Login,
			Username:  u.Login,
			FullName:  convert.FromPtr(ghUser.Name),
			Email:     u.Email,
		}
		_, err = gtClient.CreateOrGetUser(opt)
		if err != nil {
			logger.Error("failed to create user", "login", u.Login, "email", u.Email, "err", err)
			continue
		}
		logger.Info("user created or exists",
			"login", u.Login,
			"role", u.Role,
			"fullName", opt.FullName,
		)
	}
}

func main() {
	cfg := config.LoadConfig()
	logger := setupLogger(cfg.Debug)

	if cfg.Version {
		fmt.Printf("%s version %s: %s (%.7s %s)", version.App, version.Version, version.Description, version.GitCommit, version.BuildTime)
		return
	}

	if err := cfg.IsVaild(); err != nil {
		logger.Error("invalid config", "error", err)
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

	if cfg.UserListFile != "" {
		users, err := readUserList(cfg.UserListFile)
		if err != nil {
			logger.Error("failed to read user list", "error", err)
			return
		}
		createUsersFromCSV(ctx, ghClient, gtClient, users, cfg.GTSourceID, logger)
	}

	if err := migrateOrgAndRepos(ctx, cfg, logger, ghClient, gtClient); err != nil {
		logger.Error("migration failed", "error", err)
	}
}
