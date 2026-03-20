package fs

import (
	"os"
	"path/filepath"
)

func SyncDirs(src, dst string) error {
	diff, err := DiffDirs(src, dst)
	if err != nil {
		return err
	}

	for _, rel := range diff.ToCopy {
		srcPath := filepath.Join(src, rel)
		dstPath := filepath.Join(dst, rel)
		if err := CopyPath(srcPath, dstPath); err != nil {
			return err
		}
	}

	for _, rel := range diff.ToDelete {
		dstPath := filepath.Join(dst, rel)
		if err := os.Remove(dstPath); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	return nil
}

func SyncPath(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return SyncDirs(src, dst)
	}

	if _, err := os.Stat(dst); os.IsNotExist(err) {
		return CopyPath(src, dst)
	} else if err != nil {
		return err
	}

	srcHash, err := FileHash(src)
	if err != nil {
		return err
	}
	dstHash, err := FileHash(dst)
	if err != nil {
		return err
	}

	if srcHash == dstHash {
		return nil
	}

	return CopyPath(src, dst)
}
