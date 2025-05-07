# github2gitea

[English](README.md) | [简体中文](README.zh-CN.md)

[![Lint and Testing](https://github.com/appleboy/github2gitea/actions/workflows/testing.yml/badge.svg)](https://github.com/appleboy/github2gitea/actions/workflows/testing.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/appleboy/github2gitea)](https://goreportcard.com/report/github.com/appleboy/github2gitea)

一個 [CLI](https://en.wikipedia.org/wiki/Command-line_interface) 工具，用於將 [GitHub](https://github.com/) 的儲存庫（無論是組織或個人帳號）遷移到 [Gitea](https://about.gitea.com/) 平台，使用 [Golang](https://go.dev/) 開發。本專案利用 [GitHub](https://github.com/) 及 [Gitea](https://pkg.go.dev/code.gitea.io/sdk/gitea) SDK，實現無縫的儲存庫轉移。

## 目錄

- [github2gitea](#github2gitea)
  - [目錄](#目錄)
  - [使用說明](#使用說明)
    - [先決條件](#先決條件)
    - [安裝](#安裝)
    - [命令列選項](#命令列選項)
    - [範例指令](#範例指令)
    - [遷移流程](#遷移流程)
      - [使用者清單 CSV 格式](#使用者清單-csv-格式)
  - [貢獻方式](#貢獻方式)
  - [授權](#授權)

## 使用說明

### 先決條件

- 具有 `repo` 和 `admin:org` 權限的 GitHub Personal Access Token
- 具有 `write:organization` 和 `write:repository` 權限的 Gitea Personal Access Token
- Go 1.24+（若需自行編譯）

### 安裝

```bash
git clone https://github.com/appleboy/github2gitea
cd github2gitea
go build -o github2gitea cmd/github2gitea/main.go
```

### 命令列選項

| 旗標               | 說明                         | 預設值              | 必填 |
| ------------------ | ---------------------------- | ------------------- | ---- |
| `--gh-token`       | GitHub Personal Access Token | -                   | 是   |
| `--gh-skip-verify` | 跳過 GitHub TLS 驗證         | `false`             | 否   |
| `--gh-server`      | GitHub Enterprise Server URL | (public GitHub)     | 否   |
| `--gt-server`      | Gitea 伺服器 URL             | `https://gitea.com` | 否   |
| `--gt-token`       | Gitea Personal Access Token  | -                   | 是   |
| `--gt-skip-verify` | 跳過 Gitea TLS 驗證          | `false`             | 否   |
| `--gt-source-id`   | Gitea 遷移來源 ID            | `0`                 | 否   |
| `--timeout`        | 請求逾時（如 1m, 30s）       | `10m`               | 否   |
| `--source-org`     | GitHub 原始組織名稱          | -                   | 是   |
| `--target-org`     | Gitea 目標組織名稱           | -                   | 是   |
| `--debug`          | 啟用除錯日誌                 | `false`             | 否   |
| `--user-list`      | 使用者清單 CSV 檔案路徑      | -                   | 否   |

### 範例指令

基本 GitHub 到 Gitea.com 遷移：

```bash
./github2gitea \
  --gh-token your_github_token \
  --gt-token your_gitea_token \
  --source-org github-org-name \
  --target-org gitea-org-name
```

帶入使用者清單 CSV 檔案的遷移範例：

```bash
./github2gitea \
  --gh-token your_github_token \
  --gt-token your_gitea_token \
  --source-org github-org-name \
  --target-org gitea-org-name \
  --user-list users.csv
```

企業 GitHub Server 遷移：

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

### 遷移流程

1. 驗證 GitHub 與 Gitea 的身份認證
2. 在 Gitea 建立目標組織（若尚未存在）
3. 遷移所有來源 GitHub 組織的儲存庫
4. 保留儲存庫中繼資料，包括：
   - 描述
   - 可見性（公開/私有）
   - Clone URLs
   - Wiki
   - 問題（Issues）
   - 合併請求（Pull requests）
   - 發行版本（Releases）
   - 標籤（Labels）
   - 里程碑（Milestones）
5. 若有提供使用者清單 CSV 檔案，將：
   - 批次建立 Gitea 使用者帳號
   - 遷移使用者的 SSH 公鑰
   - 保留使用者角色設定
6. 針對每個儲存庫處理錯誤，並持續遷移其他儲存庫

#### 使用者清單 CSV 格式

CSV 檔案需有標頭列，且每列至少 5 個欄位，欄位順序如下：

- **created_at**（第 1 欄，建立時間，可留空）
- **id**（第 2 欄，使用者 id，可留空）
- **login**（第 3 欄，GitHub 登入名稱）
- **email**（第 4 欄，使用者 email）
- **role**（第 5 欄，使用者角色）

範例（含標頭）：

```csv
created_at,id,login,email,role
,1,alice,alice@example.com,admin
,2,bob,bob@example.com,user
```

## 貢獻方式

歡迎貢獻！如有改進建議或錯誤回報，請提出 issue 或提交 pull request。

1. Fork 本儲存庫
2. 建立新分支（`git checkout -b feature/your-feature`）
3. 提交您的更動
4. Push 到您的 fork 並開啟 pull request

## 授權

本專案採用 [MIT License](https://opensource.org/licenses/MIT) 授權。詳見 [LICENSE](LICENSE) 檔案。
