package fs

import (
	"os"
	"path/filepath"
)

type DiffResult struct {
	ToCopy   []string
	ToDelete []string
}

func DiffDirs(src, dst string) (*DiffResult, error) {
	result := &DiffResult{
		ToCopy:   make([]string, 0),
		ToDelete: make([]string, 0),
	}

	srcFiles := map[string]string{}
	dstFiles := map[string]string{}

	if err := walkFilesWithHash(src, src, srcFiles); err != nil {
		return nil, err
	}

	if _, err := os.Stat(dst); err == nil {
		if err := walkFilesWithHash(dst, dst, dstFiles); err != nil {
			return nil, err
		}
	}

	for rel, srcHash := range srcFiles {
		dstHash, ok := dstFiles[rel]
		if !ok || dstHash != srcHash {
			result.ToCopy = append(result.ToCopy, rel)
		}
	}

	for rel := range dstFiles {
		if _, ok := srcFiles[rel]; !ok {
			result.ToDelete = append(result.ToDelete, rel)
		}
	}

	return result, nil
}

func walkFilesWithHash(root, current string, out map[string]string) error {
	entries, err := os.ReadDir(current)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		absPath := filepath.Join(current, entry.Name())
		if entry.IsDir() {
			if err := walkFilesWithHash(root, absPath, out); err != nil {
				return err
			}
			continue
		}

		rel, err := filepath.Rel(root, absPath)
		if err != nil {
			return err
		}

		h, err := FileHash(absPath)
		if err != nil {
			return err
		}
		out[rel] = h
	}

	return nil
}
