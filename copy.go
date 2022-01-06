package main

import (
	"io"
	"os"
	"path"
	"path/filepath"
)

func CopyFiles(input, output string) error {
	return filepath.Walk(input, func(filePath string, info os.FileInfo, err error) error {
		if info == nil || info.IsDir() {
			return nil
		}
		if err != nil {
			return err
		}
		source, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer source.Close()

		dest, err := os.Create(path.Join(output, info.Name()))
		if err != nil {
			return err
		}
		defer dest.Close()

		_, err = io.Copy(dest, source)
		if err != nil {
			return err
		}
		return nil
	})
}
