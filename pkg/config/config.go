package config

import (
	"errors"
	"flag"

	"github.com/appleboy/com/convert"
)

// Config holds all configuration options
type Config struct {
	GHToken      string
	GHSkipVerify bool
	GHServer     string
	GTServer     string
	GTToken      string
	GTSkipVerify bool
	GTSourceID   int64
	APITimeout   string
	SourceOrg    string
	TargetOrg    string
	UserListFile string
	Debug        bool
	Version      bool
}

func (cfg *Config) IsVaild() error {
	if cfg.GHToken == "" {
		return errors.New("github token is required")
	}
	if cfg.GTToken == "" {
		return errors.New("gitea token is required")
	}
	if cfg.SourceOrg == "" {
		return errors.New("sourceOrg is required")
	}
	if cfg.TargetOrg == "" {
		return errors.New("targetOrg is required")
	}
	return nil
}

// LoadConfig parses command-line flags and returns a Config struct
func LoadConfig() *Config {
	ghToken := flag.String("gh-token", "", "GitHub Personal Access Token")
	ghSkipVerify := flag.Bool("gh-skip-verify", false, "Skip TLS verification for GitHub")
	ghServer := flag.String("gh-server", "", "GitHub Enterprise Server URL")
	gtServer := flag.String("gt-server", "https://gitea.com", "Gitea Server URL")
	gtToken := flag.String("gt-token", "", "Gitea Personal Access Token")
	gtSkipVerify := flag.Bool("gt-skip-verify", false, "Skip TLS verification for Gitea")
	gtSourceID := flag.Int64("gt-source-id", 0, "Gitea Source ID")
	apiTimeout := flag.String("timeout", "10m", "Timeout for requests")
	sourceOrg := flag.String("source-org", "", "Source organization name")
	targetOrg := flag.String("target-org", "", "Target organization name")
	userListFile := flag.String("user-list", "", "Path to user list CSV file")
	debug := flag.Bool("debug", false, "Enable debug logging")
	version := flag.Bool("version", false, "Show version information")
	flag.Parse()

	return &Config{
		GHToken:      convert.FromPtr(ghToken),
		GHSkipVerify: convert.FromPtr(ghSkipVerify),
		GHServer:     convert.FromPtr(ghServer),
		GTServer:     convert.FromPtr(gtServer),
		GTToken:      convert.FromPtr(gtToken),
		GTSkipVerify: convert.FromPtr(gtSkipVerify),
		GTSourceID:   convert.FromPtr(gtSourceID),
		APITimeout:   convert.FromPtr(apiTimeout),
		SourceOrg:    convert.FromPtr(sourceOrg),
		TargetOrg:    convert.FromPtr(targetOrg),
		UserListFile: convert.FromPtr(userListFile),
		Debug:        convert.FromPtr(debug),
		Version:      convert.FromPtr(version),
	}
}
