package git

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jesusdangerous/repomover/internal/logging"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func Push(ctx context.Context, repo *git.Repository, platform string, token string) error {

	if token == "" {
		switch platform {
		case "github":
			token = os.Getenv("GITHUB_TOKEN")
		case "gitlab":
			token = os.Getenv("GITLAB_TOKEN")
		}
		if token == "" {
			return fmt.Errorf("token is not set for platform %s", platform)
		}
	}

	var auth http.AuthMethod
	if platform == "github" {
		auth = &http.BasicAuth{
			Username: "token",
			Password: token,
		}
	} else if platform == "gitlab" {
		auth = &http.BasicAuth{
			Username: "oauth2",
			Password: token,
		}
	} else {
		return fmt.Errorf("unsupported platform: %s", platform)
	}

	options := &git.PushOptions{
		RemoteName: "origin",
		Auth:       auth,
	}

	logging.Info("Pushing changes", "platform", platform, "remote", options.RemoteName)

	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		attemptCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		err := repo.PushContext(attemptCtx, options)
		cancel()

		if err == nil || err == git.NoErrAlreadyUpToDate {
			if err == git.NoErrAlreadyUpToDate {
				logging.Info("Push skipped: already up-to-date", "platform", platform)
			} else {
				logging.Info("Push completed", "platform", platform)
			}
			return nil
		}

		lastErr = err
		logging.Warn("Push attempt failed", "attempt", attempt, "error", err.Error())
		if attempt < 3 {
			time.Sleep(2 * time.Second)
		}
	}

	return lastErr
}
