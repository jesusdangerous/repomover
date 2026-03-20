package sync

import (
	"context"
	"fmt"
	"log/slog"
	neturl "net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jesusdangerous/repomover/internal/fs"
	gitpkg "github.com/jesusdangerous/repomover/internal/git"
	"github.com/jesusdangerous/repomover/internal/logging"

	git "github.com/go-git/go-git/v5"
)

type Config struct {
	Source      string
	Target      string
	Path        string
	DestPath    string
	Commit      bool
	DryRun      bool
	Platform    string
	Token       string
	Action      string
	Incremental bool
	SourceLocal bool
	TargetLocal bool
}

func Run(cfg Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if strings.TrimSpace(cfg.Source) == "" || strings.TrimSpace(cfg.Target) == "" || strings.TrimSpace(cfg.Path) == "" {
		return fmt.Errorf("required flags are missing: --source, --target, --path")
	}

	action := strings.TrimSpace(cfg.Action)
	if action == "" {
		if cfg.Commit {
			action = "commit"
		} else {
			action = "local"
		}
	}

	if cfg.Commit && action != "commit" {
		return fmt.Errorf("conflicting flags: --commit cannot be used with --action=%s", action)
	}

	if action != "local" && action != "commit" && action != "pr" {
		return fmt.Errorf("unsupported action: %s", action)
	}

	if strings.TrimSpace(cfg.Platform) == "" {
		cfg.Platform = detectPlatform(cfg.Target)
	}

	if cfg.Platform != "github" && cfg.Platform != "gitlab" {
		return fmt.Errorf("unsupported platform: %s", cfg.Platform)
	}

	var src string
	var repo *git.Repository
	var worktreeDir string
	cleanup := make([]func(), 0)
	defer func() {
		for i := len(cleanup) - 1; i >= 0; i-- {
			cleanup[i]()
		}
	}()

	if cfg.SourceLocal {
		src = filepath.Join(cfg.Source, cfg.Path)
	} else {
		sourceDir, err := os.MkdirTemp("", "repomover-source-*")
		if err != nil {
			return err
		}
		cleanup = append(cleanup, func() { _ = os.RemoveAll(sourceDir) })
		logging.InfoAttrs(ctx, "Cloning source repository", slog.String("source", cfg.Source))
		_, err = gitpkg.CloneOrInit(cfg.Source, sourceDir)
		if err != nil {
			return err
		}
		src = filepath.Join(sourceDir, cfg.Path)
	}

	if cfg.TargetLocal {
		logging.InfoAttrs(ctx, "Using local target repository", slog.String("target", cfg.Target))
		var err error
		repo, err = git.PlainOpen(cfg.Target)
		if err != nil {
			return err
		}
		worktreeDir = cfg.Target
	} else {
		targetDir, err := os.MkdirTemp("", "repomover-target-*")
		if err != nil {
			return err
		}
		cleanup = append(cleanup, func() { _ = os.RemoveAll(targetDir) })
		logging.InfoAttrs(ctx, "Cloning target repository", slog.String("target", cfg.Target))
		repo, err = gitpkg.CloneOrInit(cfg.Target, targetDir)
		if err != nil {
			return err
		}
		worktreeDir = targetDir
	}

	dstPath := cfg.DestPath
	if dstPath == "" {
		dstPath = cfg.Path
	}
	dst := filepath.Join(worktreeDir, dstPath)

	logging.InfoAttrs(
		ctx,
		"Sync started",
		slog.String("sourcePath", src),
		slog.String("destinationPath", dst),
		slog.Bool("incremental", cfg.Incremental),
	)

	if cfg.DryRun {
		logging.InfoAttrs(ctx, "Dry run completed", slog.String("sourcePath", src), slog.String("destinationPath", dst))
		return nil
	}

	if _, err := os.Stat(src); os.IsNotExist(err) {
		// Help the user by listing top-level entries in the source repo
		entries, listErr := os.ReadDir(filepath.Dir(src))
		if listErr == nil {
			names := make([]string, 0, len(entries))
			for _, e := range entries {
				names = append(names, e.Name())
			}
			return fmt.Errorf("source path does not exist: %s\navailable entries: %v", src, names)
		}
		return fmt.Errorf("source path does not exist: %s", src)
	} else if err != nil {
		return err
	}

	var err error
	if cfg.Incremental {
		err = fs.SyncPath(src, dst)
	} else {
		err = fs.CopyPath(src, dst)
	}
	if err != nil {
		return err
	}

	hasChanges, err := gitpkg.HasChanges(repo)
	if err != nil {
		return err
	}

	if !hasChanges {
		logging.InfoCtx(ctx, "No changes detected")
		return nil
	}

	if action == "local" {
		logging.InfoAttrs(ctx, "Changes saved locally", slog.String("destinationPath", dst))
		return nil
	}

	logging.InfoCtx(ctx, "Creating commit")
	err = gitpkg.Commit(repo, "repomover sync", dstPath)
	if err != nil {
		return err
	}

	if action == "commit" {
		logging.InfoCtx(ctx, "Pushing changes")
		return gitpkg.Push(ctx, repo, cfg.Platform, cfg.Token)
	} else if action == "pr" {
		logging.InfoCtx(ctx, "Creating pull request")
		return gitpkg.CreatePR(ctx, repo, cfg.Platform, cfg.Token)
	}

	return nil
}

func detectPlatform(url string) string {
	raw := strings.ToLower(strings.TrimSpace(url))

	// SSH short syntax: git@host:owner/repo(.git)
	if strings.Contains(raw, "@") && strings.Contains(raw, ":") && !strings.Contains(raw, "://") {
		parts := strings.SplitN(raw, "@", 2)
		if len(parts) == 2 {
			hostAndPath := strings.SplitN(parts[1], ":", 2)
			if len(hostAndPath) == 2 {
				host := hostAndPath[0]
				if strings.Contains(host, "gitlab") {
					return "gitlab"
				}
				if strings.Contains(host, "github") {
					return "github"
				}
			}
		}
	}
	parsed, err := neturl.Parse(raw)
	if err == nil {
		host := strings.ToLower(parsed.Host)
		if strings.Contains(host, "gitlab") {
			return "gitlab"
		}
		if strings.Contains(host, "github") {
			return "github"
		}
	}

	if strings.Contains(raw, "gitlab") {
		return "gitlab"
	}
	return "github"
}
