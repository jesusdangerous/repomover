package git

import (
	"fmt"

	"github.com/go-git/go-git/v5/plumbing"

	git "github.com/go-git/go-git/v5"
)

func CreateBranch(repo *git.Repository, branchName string) error {
	if branchName == "" {
		return fmt.Errorf("branch name cannot be empty")
	}

	head, err := repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}

	newRef := plumbing.NewHashReference(
		plumbing.NewBranchReferenceName(branchName),
		head.Hash(),
	)
	_ = repo.Storer.SetReference(newRef)

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branchName),
		Force:  true,
	})
	if err != nil {
		return fmt.Errorf("failed to checkout branch %s: %w", branchName, err)
	}

	return nil
}
