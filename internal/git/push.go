package git

import (
	"context"
	"time"

	"github.com/jesusdangerous/repomover/internal/logging"

	git "github.com/go-git/go-git/v5"
)

func Push(ctx context.Context, repo *git.Repository, cfg AuthConfig) error {
	remoteURL, err := getRemoteURL(repo)
	if err != nil {
		return err
	}

	auth, authMode, err := authForRemote(remoteURL, cfg, true)
	if err != nil {
		return err
	}

	options := &git.PushOptions{
		RemoteName: "origin",
		Auth:       auth,
	}

	logging.Info("Pushing changes", "platform", cfg.Platform, "remote", options.RemoteName, "auth", authMode)

	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		attemptCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		err := repo.PushContext(attemptCtx, options)
		cancel()

		if err == nil || err == git.NoErrAlreadyUpToDate {
			if err == git.NoErrAlreadyUpToDate {
				logging.Info("Push skipped: already up-to-date", "platform", cfg.Platform)
			} else {
				logging.Info("Push completed", "platform", cfg.Platform)
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
