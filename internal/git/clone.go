package git

import (
	"errors"
	"os"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

func CloneOrInit(url string, dir string) (*git.Repository, error) {

	repo, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL: url,
	})

	if err == nil {
		return repo, nil
	}

	if !errors.Is(err, transport.ErrEmptyRemoteRepository) {
		return nil, err
	}

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, err
	}

	repo, err = git.PlainInit(dir, false)
	if err != nil {
		return nil, err
	}

	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{url},
	})
	if err != nil {
		return nil, err
	}

	return repo, nil
}
