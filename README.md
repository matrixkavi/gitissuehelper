# Git Issue Helper

A Go CLI tool to create multiple GitHub issues with the same description across repositories in an organization.

## Features

- Create issues across multiple repositories in an organization
- Automatically scan all repos if specific repos are not provided
- Optional label support
- Uses GitHub REST API with OAuth2 authentication
- Progress tracking with success/failure summary

## Building

```bash
go mod download
go build -o gitissuehelper
```

## Usage

```bash
./gitissuehelper create --org <org> --title <title> --description <description> [--repos <repos>] [--labels <labels>] [--token <token>]
```

### Options

- `--org, -o` - GitHub organization name (required)
- `--title, -t` - Issue title (required)
- `--description, -d` - Issue description (required)
- `--repos, -r` - Comma-separated list of repository names (optional; if omitted, all repos in org are used)
- `--labels, -l` - Comma-separated labels to add to issues (optional)
- `--token` - GitHub API token (optional; uses `GITHUB_TOKEN` env var if not provided)

### Examples

Create issues in all repositories:
```bash
export GITHUB_TOKEN=your_token_here
./gitissuehelper create --org myorg --title "Update docs" --description "Please update documentation"
```

Create issues in specific repositories with labels:
```bash
./gitissuehelper create --org myorg --repos repo1,repo2,repo3 --title "Update docs" --description "Please update documentation" --labels "documentation,help-wanted"
```

## Authentication

The tool requires a GitHub API token for authentication. You can provide it in two ways:

1. Set the `GITHUB_TOKEN` environment variable
2. Pass it using the `-token` flag

To create a personal access token:
1. Go to GitHub Settings → Developer settings → Personal access tokens
2. Create a new token with `repo` scope
3. Use the token with this tool

## Requirements

- Go 1.21 or higher
- GitHub API token with `repo` scope
