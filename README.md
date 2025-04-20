# github2gitea

[![Lint and Testing](https://github.com/appleboy/github2gitea/actions/workflows/testing.yml/badge.svg)](https://github.com/appleboy/github2gitea/actions/workflows/testing.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/appleboy/github2gitea)](https://goreportcard.com/report/github.com/appleboy/github2gitea)

A [CLI](https://en.wikipedia.org/wiki/Command-line_interface) tool to migrate [GitHub](https://github.com/) repositories (from organizations or personal accounts) to a [Gitea](https://about.gitea.com/) platform, built with [Golang](https://go.dev/). This project leverages the [GitHub](https://github.com/) and [Gitea](https://pkg.go.dev/code.gitea.io/sdk/gitea) SDKs for seamless repository transfer.

## Usage

### Prerequisites

- GitHub Personal Access Token with `repo` and `admin:org` scopes
- Gitea Personal Access Token with `write:organization` and `write:repository` permissions
- Go 1.20+ (if building from source)

### Installation

```bash
git clone https://github.com/appleboy/github2gitea
cd github2gitea
go build -o github2gitea cmd/github2gitea/main.go
```

### Command-Line Options

| Flag               | Description                      | Default             | Required |
| ------------------ | -------------------------------- | ------------------- | -------- |
| `--gh-token`       | GitHub Personal Access Token     | -                   | Yes      |
| `--gh-skip-verify` | Skip TLS verification for GitHub | `false`             | No       |
| `--gh-server`      | GitHub Enterprise Server URL     | (public GitHub)     | No       |
| `--gt-server`      | Gitea Server URL                 | `https://gitea.com` | No       |
| `--gt-token`       | Gitea Personal Access Token      | -                   | Yes      |
| `--gt-skip-verify` | Skip TLS verification for Gitea  | `false`             | No       |
| `--gt-source-id`   | Gitea Migration Source ID        | `0`                 | No       |
| `--timeout`        | Request timeout (e.g., 1m, 30s)  | `10m`               | No       |
| `--source-org`     | Source GitHub organization name  | -                   | Yes      |
| `--target-org`     | Target Gitea organization name   | -                   | Yes      |
| `--debug`          | Enable debug logging             | `false`             | No       |

### Example Commands

Basic migration from GitHub to Gitea.com:

```bash
./github2gitea \
  --gh-token your_github_token \
  --gt-token your_gitea_token \
  --source-org github-org-name \
  --target-org gitea-org-name
```

Enterprise GitHub Server migration:

```bash
./github2gitea \
  --gh-server https://github.example.com \
  --gh-token your_github_token \
  --gt-server https://gitea.example.com \
  --gt-token your_gitea_token \
  --source-org enterprise-org \
  --target-org new-gitea-org \
  --timeout 5m \
  --debug
```

### Migration Process

1. Validates authentication with GitHub and Gitea
2. Creates target organization in Gitea (if not exists)
3. Migrates all repositories from source GitHub organization
4. Preserves repository metadata including:
   - Description
   - Visibility (public/private)
   - Clone URLs
5. Handles errors per-repository while continuing migration

## Contributing

Contributions are welcome! Please open issues or submit pull requests for improvements and bug fixes.

1. Fork the repository
2. Create a new branch (`git checkout -b feature/your-feature`)
3. Commit your changes
4. Push to your fork and open a pull request

## License

This project is licensed under the [MIT License](https://opensource.org/licenses/MIT). See the [LICENSE](LICENSE) file for details.
