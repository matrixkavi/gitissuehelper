package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/go-github/v57/github"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

type IssueCreator struct {
	client *github.Client
	ctx    context.Context
	org    string
	title  string
	desc   string
	labels []string
}

// NewIssueCreator creates a new IssueCreator instance
func NewIssueCreator(token, org, title, desc string, labels []string) (*IssueCreator, error) {
	if token == "" {
		return nil, fmt.Errorf("GitHub token is required. Set GITHUB_TOKEN env var or use --token flag")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return &IssueCreator{
		client: client,
		ctx:    ctx,
		org:    org,
		title:  title,
		desc:   desc,
		labels: labels,
	}, nil
}

// GetAllRepositories fetches all repositories for an organization
func (ic *IssueCreator) GetAllRepositories() ([]string, error) {
	opts := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var repos []string
	for {
		repoList, resp, err := ic.client.Repositories.ListByOrg(ic.ctx, ic.org, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch repositories: %w", err)
		}

		for _, repo := range repoList {
			repos = append(repos, *repo.Name)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return repos, nil
}

// CreateIssue creates an issue in a specific repository
func (ic *IssueCreator) CreateIssue(repo string) error {
	issueRequest := &github.IssueRequest{
		Title:  &ic.title,
		Body:   &ic.desc,
		Labels: &ic.labels,
	}

	_, _, err := ic.client.Issues.Create(ic.ctx, ic.org, repo, issueRequest)
	if err != nil {
		return fmt.Errorf("failed to create issue in %s/%s: %w", ic.org, repo, err)
	}

	return nil
}

// CreateIssuesInRepositories creates issues in multiple repositories
func (ic *IssueCreator) CreateIssuesInRepositories(repos []string) (int, int) {
	success := 0
	failed := 0

	for _, repo := range repos {
		fmt.Printf("Creating issue in %s/%s... ", ic.org, repo)
		if err := ic.CreateIssue(repo); err != nil {
			fmt.Printf("✗ (%v)\n", err)
			failed++
		} else {
			fmt.Println("✓")
			success++
		}
	}

	return success, failed
}

var rootCmd = &cobra.Command{
	Use:   "gitissuehelper",
	Short: "Create GitHub issues across multiple repositories",
	Long: `gitissuehelper is a CLI tool to create issues across multiple repositories in a GitHub organization.
It supports batch issue creation with customizable titles, descriptions, and labels.`,
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create issues in repositories",
	RunE:  runCreate,
}

func runCreate(cmd *cobra.Command, args []string) error {
	// Get configuration from flags and environment
	org := viper.GetString("org")
	title := viper.GetString("title")
	desc := viper.GetString("description")
	repos := viper.GetString("repos")
	labels := viper.GetString("labels")
	token := viper.GetString("token")

	// Validate required flags
	if org == "" || title == "" || desc == "" {
		return fmt.Errorf("missing required arguments: --org, --title, and --description are required")
	}

	// Get GitHub token
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}

	// Create IssueCreator
	labelList := []string{}
	if labels != "" {
		labelList = strings.Split(labels, ",")
		// Trim whitespace from labels
		for i := range labelList {
			labelList[i] = strings.TrimSpace(labelList[i])
		}
	}

	creator, err := NewIssueCreator(token, org, title, desc, labelList)
	if err != nil {
		return fmt.Errorf("failed to initialize: %v", err)
	}

	// Get repositories
	var repoList []string
	if repos != "" {
		// Use provided repositories
		repoList = strings.Split(repos, ",")
		// Trim whitespace
		for i := range repoList {
			repoList[i] = strings.TrimSpace(repoList[i])
		}
	} else {
		// Fetch all repositories
		fmt.Printf("Fetching repositories from organization: %s...\n", org)
		repoList, err = creator.GetAllRepositories()
		if err != nil {
			return fmt.Errorf("failed to fetch repositories: %v", err)
		}
	}

	if len(repoList) == 0 {
		return fmt.Errorf("no repositories found")
	}

	// Create issues
	fmt.Printf("Creating issues in organization: %s\n", org)
	fmt.Printf("Title: %s\n", title)
	fmt.Printf("Repositories: %d\n", len(repoList))
	fmt.Println("---")

	success, failed := creator.CreateIssuesInRepositories(repoList)

	fmt.Println("---")
	fmt.Printf("Summary: %d succeeded, %d failed\n", success, failed)

	if failed > 0 {
		os.Exit(1)
	}

	return nil
}

func init() {
	// Bind environment variables
	viper.SetEnvPrefix("GITISSUEHELPER")
	viper.AutomaticEnv()

	// Create command flags
	createCmd.Flags().StringP("org", "o", "", "GitHub organization name (required)")
	createCmd.Flags().StringP("title", "t", "", "Issue title (required)")
	createCmd.Flags().StringP("description", "d", "", "Issue description (required)")
	createCmd.Flags().StringP("repos", "r", "", "Comma-separated list of repository names (optional; if omitted, all repos in org are used)")
	createCmd.Flags().StringP("labels", "l", "", "Comma-separated labels to add to issues (optional)")
	createCmd.Flags().String("token", "", "GitHub API token (optional; uses GITHUB_TOKEN env var if not provided)")

	// Bind flags to Viper
	viper.BindPFlag("org", createCmd.Flags().Lookup("org"))
	viper.BindPFlag("title", createCmd.Flags().Lookup("title"))
	viper.BindPFlag("description", createCmd.Flags().Lookup("description"))
	viper.BindPFlag("repos", createCmd.Flags().Lookup("repos"))
	viper.BindPFlag("labels", createCmd.Flags().Lookup("labels"))
	viper.BindPFlag("token", createCmd.Flags().Lookup("token"))

	// Add commands
	rootCmd.AddCommand(createCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
