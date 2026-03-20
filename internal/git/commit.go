package git

import (
	"path/filepath"
	"strings"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func Commit(repo *git.Repository, message string, paths ...string) error {

	w, err := repo.Worktree()
	if err != nil {
		return err
	}

	status, err := w.Status()
	if err != nil {
		return err
	}

	if len(paths) == 0 {
		if err := w.AddGlob("."); err != nil {
			return err
		}
		for path, st := range status {
			if st.Worktree == git.Deleted || st.Staging == git.Deleted {
				if _, err := w.Remove(path); err != nil {
					return err
				}
			}
		}
	} else {
		normalized := make([]string, 0, len(paths))
		for _, p := range paths {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			normalized = append(normalized, filepath.Clean(p))
		}

		for path, st := range status {
			if !matchesAnyPath(path, normalized) {
				continue
			}

			if st.Worktree == git.Deleted || st.Staging == git.Deleted {
				if _, err := w.Remove(path); err != nil {
					return err
				}
				continue
			}

			if _, err := w.Add(path); err != nil {
				return err
			}
		}

		for _, p := range normalized {
			if p == "" || p == "." {
				if err := w.AddGlob("."); err != nil {
					return err
				}
				continue
			}
		}
	}

	_, err = w.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "repomover",
			Email: "bot@repomover",
			When:  time.Now(),
		},
	})

	return err
}

func matchesAnyPath(changedPath string, roots []string) bool {
	for _, root := range roots {
		if root == "." {
			return true
		}
		if changedPath == root {
			return true
		}
		prefix := root + string(filepath.Separator)
		if strings.HasPrefix(changedPath, prefix) {
			return true
		}
	}
	return false
}
