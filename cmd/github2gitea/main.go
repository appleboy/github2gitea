package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"time"

	gt "github.com/appleboy/github2gitea/pkg/gitea"
	gh "github.com/appleboy/github2gitea/pkg/github"
	"github.com/appleboy/github2gitea/pkg/migrate"

	"github.com/appleboy/com/convert"
)

func main() {
	ghToken := flag.String("gh-token", "", "GitHub Personal Access Token")
	ghSkipVerify := flag.Bool("gh-skip-verify", false, "Skip TLS verification for GitHub")
	ghServer := flag.String("gh-server", "", "GitHub Enterprise Server URL")
	gtServer := flag.String("gt-server", "https://gitea.com", "Gitea Server URL")
	gtToken := flag.String("gt-token", "", "Gitea Personal Access Token")
	gtSkipVerify := flag.Bool("gt-skip-verify", false, "Skip TLS verification for Gitea")
	gtSourceID := flag.Int64("gt-source-id", 0, "Gitea Source ID")
	apiTimeout := flag.String("timeout", "1m", "Timeout for requests")
	sourceOrg := flag.String("source-org", "", "Source organization name")
	targetOrg := flag.String("target-org", "", "Target organization name")
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	logLevel := slog.LevelInfo
	if convert.FromPtr(debug) {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(log.Writer(), &slog.HandlerOptions{
		Level: logLevel,
	}))

	sourceOrgName := convert.FromPtr(sourceOrg)
	targetOrgName := convert.FromPtr(targetOrg)

	if sourceOrgName == "" || targetOrgName == "" {
		logger.Error("source or target org is empty")
		return
	}

	// check timeout format
	timeout, err := time.ParseDuration(convert.FromPtr(apiTimeout))
	if err != nil {
		logger.Error("failed to parse timeout", "error", err)
		return
	}
	// command timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ghClient, err := gh.NewClient(&gh.Config{
		Token:      convert.FromPtr(ghToken),
		Server:     convert.FromPtr(ghServer),
		SkipVerify: convert.FromPtr(ghSkipVerify),
		Logger:     logger,
	})
	if err != nil {
		logger.Error("failed to create GitHub client", "error", err)
		return
	}

	gtClient, err := gt.New(ctx, &gt.Config{
		Server:     convert.FromPtr(gtServer),
		Token:      convert.FromPtr(gtToken),
		SkipVerify: convert.FromPtr(gtSkipVerify),
		Logger:     logger,
		SourceID:   convert.FromPtr(gtSourceID),
	})
	if err != nil {
		logger.Error("failed to create gitea client", "error", err)
		return
	}

	// get github current user
	ghUser, err := ghClient.GetCurrentUser(ctx)
	if err != nil {
		logger.Error("failed to get current github user", "error", err)
		return
	}
	logger.Info("github user", "login", convert.FromPtr(ghUser.Login))
	logger.Info("github user", "name", convert.FromPtr(ghUser.Name))
	logger.Info("github user", "email", convert.FromPtr(ghUser.Email))

	// get gitea current user
	gtUser, err := gtClient.GetCurrentUser()
	if err != nil {
		logger.Error("failed to get current gitea user", "error", err)
		return
	}
	logger.Info("gitea user", "login", gtUser.UserName)
	logger.Info("gitea user", "name", gtUser.FullName)
	logger.Info("gitea user", "email", gtUser.Email)

	// get github organization
	ghOrg, err := ghClient.GetOrg(ctx, sourceOrgName)
	if err != nil {
		logger.Error("failed to get github org", "error", err)
		return
	}

	m := migrate.New(
		ghClient,
		gtClient,
		logger,
	)

	// create new gitea organization
	_, err = m.CreateNewOrg(ctx, migrate.CreateNewOrgOption{
		Name:        targetOrgName,
		Description: convert.FromPtr(ghOrg.Description),
		Public:      false,
		SoucreID:    convert.FromPtr(gtSourceID),
	})
	if err != nil {
		logger.Error("failed to create gitea org", "error", err)
		return
	}

	// get github repo list from organization
	ghRepos, err := ghClient.ListOrgRepos(ctx, sourceOrgName)
	if err != nil {
		logger.Error("failed to get github org repos", "error", err)
		return
	}

	for _, repo := range ghRepos {
		logger.Info("github repo",
			"name", convert.FromPtr(repo.Name),
			"full_name", convert.FromPtr(repo.FullName),
		)
	}
}
