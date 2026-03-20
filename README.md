# repomover

`repomover` syncs files or directories from one repository to another.

It supports:
- local and remote repositories
- full copy or incremental sync by content hash
- optional commit/push/PR flow
- GitHub and GitLab for remote operations
- HTTPS token and SSH auth for Git clone/push

Authentication options:

- HTTPS + token:
	- GitHub: `GITHUB_TOKEN`
	- GitLab: `GITLAB_TOKEN`
- SSH:
	- SSH agent, or
	- `REPOMOVER_SSH_KEY_PATH` (+ optional `REPOMOVER_SSH_KEY_PASSPHRASE`)
	- optional `REPOMOVER_SSH_USER` (default: `git`)
	- CLI flags `--ssh-key-path`, `--ssh-key-passphrase`, `--ssh-user` (these override env vars)

## Quick Start

Preview sync without writing changes:

```bash
./repomover --source https://github.com/org/source.git --target https://github.com/org/target.git --path services/api --dry-run
```

Sync from local repo to local repo:

```bash
./repomover --source /repos/source --source-local --target /repos/target --target-local --path services/api --action local
```

Sync and push commit:

```bash
./repomover --source https://github.com/org/source.git --target https://github.com/org/target.git --path services/api --action commit --platform github
```

Sync and push commit over SSH:

```bash
./repomover --source git@github.com:org/source.git --target git@github.com:org/target.git --path services/api --action commit
```

Sync, push and create PR:

```bash
./repomover --source https://gitlab.com/group/source.git --target https://gitlab.com/group/target.git --path services/api --action pr --platform gitlab
```

## Actions

- `local`: copy/sync changes only in local worktree (no commit, no push)
- `commit`: copy/sync + commit + push
- `pr`: copy/sync + commit + push + create pull/merge request

Legacy flag `--commit` is still supported and treated as `--action=commit`.

## Flags

| Flag | Description |
| --- | --- |
| `--source` | Source repository URL or local path (required) |
| `--target` | Target repository URL or local path (required) |
| `--path` | Source path inside source repo (required) |
| `--dest-path` | Destination path inside target repo (default: same as `--path`) |
| `--source-local` | Treat `--source` as local filesystem path |
| `--target-local` | Treat `--target` as local filesystem path |
| `--action` | `local`, `commit`, `pr` |
| `--commit` | Legacy alias for `--action=commit` |
| `--dry-run` | Validate and print plan without writing changes |
| `--incremental` | Copy only changed files and delete removed files |
| `--platform` | `github` or `gitlab` (auto-detected for non-local actions) |
| `--token` | Explicit token for push/PR (otherwise from env var) |
| `--ssh-user` | SSH username for git auth (default: `git`) |
| `--ssh-key-path` | Path to SSH private key for git auth |
| `--ssh-key-passphrase` | Passphrase for SSH private key |
| `--log-level` | Log verbosity: debug, info, warn, error (default: warn) |

## Output

By default, only the result summary is displayed on stdout:

For `--action=local` (or no action):
```
✓ Synced files to /repos/target/services/api
```

For `--action=commit`:
```
✓ Changes committed and pushed
```

For `--action=pr`:
```
✓ Pull request created: repomover-sync -> main
```

For `--dry-run`:
```
✓ Validation successful
  Source: /source/path
  Destination: /dest/path
```

### SSH authentication issues

- **SSH agent not working**: set `REPOMOVER_SSH_KEY_PATH=/path/to/id_rsa` and optionally `REPOMOVER_SSH_KEY_PASSPHRASE`
- **Or via CLI**: `--ssh-key-path /home/user/.ssh/id_rsa --ssh-key-passphrase "if needed"`

### PR creation fails with "No commits between branches"

- This error means source and target have identical content
- Check: is there actual content to sync in `--path`?
- Check: is source different from target?

### Token/API errors

- GitHub: use `GITHUB_TOKEN` or `--token ghp_...`
- GitLab: use `GITLAB_TOKEN` or `--token glpat-...`
- Token must have permissions for push and PR/MR creation
