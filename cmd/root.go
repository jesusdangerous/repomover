package cmd

import (
	"strings"

	"github.com/jesusdangerous/repomover/internal/logging"
	syncpkg "github.com/jesusdangerous/repomover/internal/sync"
	"github.com/spf13/cobra"
)

var cfg syncpkg.Config
var logLevel string

var rootCmd = &cobra.Command{
	Use:   "repomover",
	Short: "Repository synchronization tool",
	RunE: func(cmd *cobra.Command, args []string) error {
		logging.Init(logLevel)

		cfg.Source = strings.TrimSpace(cfg.Source)
		cfg.Target = strings.TrimSpace(cfg.Target)
		cfg.Path = strings.TrimSpace(cfg.Path)
		cfg.DestPath = strings.TrimSpace(cfg.DestPath)
		cfg.Platform = strings.TrimSpace(strings.ToLower(cfg.Platform))
		cfg.Token = strings.TrimSpace(cfg.Token)
		cfg.Action = strings.TrimSpace(strings.ToLower(cfg.Action))
		cfg.SSHUser = strings.TrimSpace(cfg.SSHUser)
		cfg.SSHKeyPath = strings.TrimSpace(cfg.SSHKeyPath)
		cfg.SSHPassphrase = strings.TrimSpace(cfg.SSHPassphrase)

		return syncpkg.Run(cfg)
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	flags := rootCmd.Flags()

	flags.StringVar(&cfg.Source, "source", "", "Source repository URL or local path")
	flags.StringVar(&cfg.Target, "target", "", "Target repository URL or local path")
	flags.StringVar(&cfg.Path, "path", "", "Path to sync from source")
	flags.StringVar(&cfg.DestPath, "dest-path", "", "Destination path in target repository (default: same as --path)")

	flags.BoolVar(&cfg.Commit, "commit", false, "Legacy flag for commit action (same as --action=commit)")
	flags.StringVar(&cfg.Action, "action", "local", "Action after sync: local, commit, pr")
	flags.BoolVar(&cfg.DryRun, "dry-run", false, "Show what would be done without writing changes")
	flags.BoolVar(&cfg.Incremental, "incremental", false, "Sync only changed files")

	flags.StringVar(&cfg.Platform, "platform", "", "Git platform: github or gitlab (auto-detected by default)")
	flags.StringVar(&cfg.Token, "token", "", "API token for push/PR operations")
	flags.StringVar(&cfg.SSHUser, "ssh-user", "", "SSH username for git auth (default: git)")
	flags.StringVar(&cfg.SSHKeyPath, "ssh-key-path", "", "Path to SSH private key for git auth")
	flags.StringVar(&cfg.SSHPassphrase, "ssh-key-passphrase", "", "Passphrase for SSH private key")
	flags.BoolVar(&cfg.SourceLocal, "source-local", false, "Treat --source as local filesystem path")
	flags.BoolVar(&cfg.TargetLocal, "target-local", false, "Treat --target as local repository path")

	flags.StringVar(&logLevel, "log-level", "info", "Log level: debug, info, warn, error")
}
