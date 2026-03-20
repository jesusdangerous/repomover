package git

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/jesusdangerous/repomover/internal/logging"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type GitHubPR struct {
	Title string `json:"title"`
	Head  string `json:"head"`
	Base  string `json:"base"`
	Body  string `json:"body"`
}

type GitLabPR struct {
	Title        string `json:"title"`
	SourceBranch string `json:"source_branch"`
	TargetBranch string `json:"target_branch"`
	Description  string `json:"description"`
}

func CreatePR(ctx context.Context, repo *git.Repository, platform string, token string) error {

	if token == "" {
		if platform == "github" {
			token = os.Getenv("GITHUB_TOKEN")
		} else if platform == "gitlab" {
			token = os.Getenv("GITLAB_TOKEN")
		}
		if token == "" {
			return fmt.Errorf("token is not set")
		}
	}

	head, err := repo.Head()
	if err != nil {
		return err
	}
	headBranch := head.Name().Short()

	baseBranch := detectBaseBranch(repo)

	remoteURL, err := getRemoteURL(repo)
	if err != nil {
		return err
	}

	owner, repoName, err := parseOwnerRepo(remoteURL)
	if err != nil {
		return err
	}

	var apiURL string
	var payload interface{}

	if platform == "github" {
		apiURL = fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls", owner, repoName)
		payload = GitHubPR{
			Title: "repomover sync",
			Head:  headBranch,
			Base:  baseBranch,
			Body:  "Automated sync by repomover",
		}
	} else if platform == "gitlab" {
		apiURL = fmt.Sprintf("https://gitlab.com/api/v4/projects/%s%%2F%s/merge_requests", owner, repoName)
		payload = GitLabPR{
			Title:        "repomover sync",
			SourceBranch: headBranch,
			TargetBranch: baseBranch,
			Description:  "Automated sync by repomover",
		}
	} else {
		return fmt.Errorf("unsupported platform: %s", platform)
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create PR: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	logging.InfoAttrs(
		ctx,
		"Pull request created successfully",
		slog.String("platform", platform),
		slog.String("head", headBranch),
		slog.String("base", baseBranch),
		slog.String("repo", owner+"/"+repoName),
	)
	return nil
}

func detectBaseBranch(repo *git.Repository) string {
	if _, err := repo.Reference(plumbing.NewBranchReferenceName("main"), true); err == nil {
		return "main"
	}
	if _, err := repo.Reference(plumbing.NewBranchReferenceName("master"), true); err == nil {
		return "master"
	}
	return "main"
}

func getRemoteURL(repo *git.Repository) (string, error) {
	origin, err := repo.Remote("origin")
	if err == nil && len(origin.Config().URLs) > 0 {
		return origin.Config().URLs[0], nil
	}

	remotes, err := repo.Remotes()
	if err != nil {
		return "", err
	}
	if len(remotes) == 0 || len(remotes[0].Config().URLs) == 0 {
		return "", fmt.Errorf("no remotes found")
	}

	return remotes[0].Config().URLs[0], nil
}

func parseOwnerRepo(remoteURL string) (string, string, error) {
	trimmed := strings.TrimSpace(strings.TrimSuffix(remoteURL, ".git"))

	if strings.Contains(trimmed, "@") && strings.Contains(trimmed, ":") && !strings.Contains(trimmed, "://") {
		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid remote URL: %s", remoteURL)
		}
		pathParts := strings.Split(strings.Trim(parts[1], "/"), "/")
		if len(pathParts) < 2 {
			return "", "", fmt.Errorf("invalid remote URL path: %s", remoteURL)
		}
		return pathParts[len(pathParts)-2], pathParts[len(pathParts)-1], nil
	}

	u, err := url.Parse(trimmed)
	if err != nil {
		return "", "", fmt.Errorf("invalid remote URL: %w", err)
	}

	pathParts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(pathParts) < 2 {
		return "", "", fmt.Errorf("invalid remote URL path: %s", remoteURL)
	}

	return pathParts[len(pathParts)-2], pathParts[len(pathParts)-1], nil
}
