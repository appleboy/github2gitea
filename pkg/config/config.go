package config

import (
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
	Debug        bool
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
	debug := flag.Bool("debug", false, "Enable debug logging")
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
		Debug:        convert.FromPtr(debug),
	}
}
