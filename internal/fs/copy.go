package fs

import (
	"io"
	"os"
	"path/filepath"
)

func CopyDir(src string, dst string) error {

	if err := os.MkdirAll(dst, os.ModePerm); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, e := range entries {

		srcPath := filepath.Join(src, e.Name())
		dstPath := filepath.Join(dst, e.Name())

		if e.IsDir() {
			if err := CopyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func CopyPath(src string, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return CopyDir(src, dst)
	}

	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, os.ModePerm); err != nil {
		return err
	}
	return CopyFile(src, dst)
}

func CopyFile(src, dst string) error {

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)

	return err
}
