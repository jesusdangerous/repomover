package git

import (
	git "github.com/go-git/go-git/v5"
)

func HasChanges(repo *git.Repository) (bool, error) {

	w, err := repo.Worktree()
	if err != nil {
		return false, err
	}

	status, err := w.Status()
	if err != nil {
		return false, err
	}

	return !status.IsClean(), nil
}
