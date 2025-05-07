# github2gitea

[English](README.md) | [繁體中文](README.zh-TW.md)

[![Lint and Testing](https://github.com/appleboy/github2gitea/actions/workflows/testing.yml/badge.svg)](https://github.com/appleboy/github2gitea/actions/workflows/testing.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/appleboy/github2gitea)](https://goreportcard.com/report/github.com/appleboy/github2gitea)

一个 [CLI](https://en.wikipedia.org/wiki/Command-line_interface) 工具，用于将 [GitHub](https://github.com/) 的仓库（无论是组织还是个人账号）迁移到 [Gitea](https://about.gitea.com/) 平台，使用 [Golang](https://go.dev/) 开发。本项目利用 [GitHub](https://github.com/) 及 [Gitea](https://pkg.go.dev/code.gitea.io/sdk/gitea) SDK，实现无缝的仓库迁移。

## 目录

- [github2gitea](#github2gitea)
  - [目录](#目录)
  - [使用说明](#使用说明)
    - [先决条件](#先决条件)
    - [安装](#安装)
    - [命令行选项](#命令行选项)
    - [示例命令](#示例命令)
    - [迁移流程](#迁移流程)
      - [用户清单 CSV 格式](#用户清单-csv-格式)
  - [贡献方式](#贡献方式)
  - [许可证](#许可证)

## 使用说明

### 先决条件

- 具有 `repo` 和 `admin:org` 权限的 GitHub Personal Access Token
- 具有 `write:organization` 和 `write:repository` 权限的 Gitea Personal Access Token
- Go 1.24+（如需自行编译）

### 安装

```bash
git clone https://github.com/appleboy/github2gitea
cd github2gitea
go build -o github2gitea cmd/github2gitea/main.go
```

### 命令行选项

| 标志               | 说明                         | 默认值              | 必填 |
| ------------------ | ---------------------------- | ------------------- | ---- |
| `--gh-token`       | GitHub Personal Access Token | -                   | 是   |
| `--gh-skip-verify` | 跳过 GitHub TLS 验证         | `false`             | 否   |
| `--gh-server`      | GitHub Enterprise Server URL | (public GitHub)     | 否   |
| `--gt-server`      | Gitea 服务器 URL             | `https://gitea.com` | 否   |
| `--gt-token`       | Gitea Personal Access Token  | -                   | 是   |
| `--gt-skip-verify` | 跳过 Gitea TLS 验证          | `false`             | 否   |
| `--gt-source-id`   | Gitea 迁移来源 ID            | `0`                 | 否   |
| `--timeout`        | 请求超时（如 1m, 30s）       | `10m`               | 否   |
| `--source-org`     | GitHub 源组织名称            | -                   | 是   |
| `--target-org`     | Gitea 目标组织名称           | -                   | 是   |
| `--debug`          | 启用调试日志                 | `false`             | 否   |
| `--user-list`      | 用户清单 CSV 文件路径        | -                   | 否   |

### 示例命令

基本 GitHub 到 Gitea.com 迁移：

```bash
./github2gitea \
  --gh-token your_github_token \
  --gt-token your_gitea_token \
  --source-org github-org-name \
  --target-org gitea-org-name
```

带用户清单 CSV 文件的迁移示例：

```bash
./github2gitea \
  --gh-token your_github_token \
  --gt-token your_gitea_token \
  --source-org github-org-name \
  --target-org gitea-org-name \
  --user-list users.csv
```

企业 GitHub Server 迁移：

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

### 迁移流程

1. 验证 GitHub 与 Gitea 的身份认证
2. 在 Gitea 创建目标组织（如尚未存在）
3. 迁移所有源 GitHub 组织的仓库
4. 保留仓库元数据，包括：
   - 描述
   - 可见性（公开/私有）
   - Clone URLs
   - Wiki
   - 问题（Issues）
   - 拉取请求（Pull requests）
   - 发布（Releases）
   - 标签（Labels）
   - 里程碑（Milestones）
5. 如果提供了用户清单 CSV 文件，将：
   - 批量创建 Gitea 用户账号
   - 迁移用户的 SSH 公钥
   - 保留用户角色设置
6. 针对每个仓库处理错误，并继续迁移其他仓库

#### 用户清单 CSV 格式

CSV 文件需有表头行，且每行至少 5 个字段，字段顺序如下：

- **created_at**（第 1 列，创建时间，可为空）
- **id**（第 2 列，用户 id，可为空）
- **login**（第 3 列，GitHub 登录名）
- **email**（第 4 列，用户邮箱）
- **role**（第 5 列，用户角色）

示例（含表头）：

```csv
created_at,id,login,email,role
,1,alice,alice@example.com,admin
,2,bob,bob@example.com,user
```

## 贡献方式

欢迎贡献！如有改进建议或错误反馈，请提交 issue 或 pull request。

1. Fork 本仓库
2. 创建新分支（`git checkout -b feature/your-feature`）
3. 提交您的更改
4. Push 到您的 fork 并创建 pull request

## 许可证

本项目采用 [MIT License](https://opensource.org/licenses/MIT) 授权。详见 [LICENSE](LICENSE) 文件。
