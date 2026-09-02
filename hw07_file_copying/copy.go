package main

import (
	"errors"
	"io"
	"os"
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
	ErrIsDirectory           = errors.New("is directory")
	ErrEmptyPath             = errors.New("empty path")
)

func Copy(fromPath, toPath string, offset, limit int64) error {
	if fromPath == "" || toPath == "" {
		return ErrEmptyPath
	}

	fileIn, err := os.Open(fromPath)
	if err != nil {
		return err
	}
	defer func() { _ = fileIn.Close() }()

	info, err := fileIn.Stat()
	if err != nil {
		return err
	}

	if info.IsDir() {
		return ErrIsDirectory
	}

	if info.Size() == 0 {
		return ErrUnsupportedFile
	}

	if offset > 0 && offset > info.Size() {
		return ErrOffsetExceedsFileSize
	}

	if _, err = fileIn.Seek(offset, io.SeekStart); err != nil {
		return err
	}

	fileCopy, err := os.Create(toPath)
	if err != nil {
		return err
	}
	defer func() { _ = fileCopy.Close() }()

	if limit == 0 {
		_, err = io.Copy(fileCopy, fileIn)
	} else {
		_, err = io.CopyN(fileCopy, fileIn, limit)
	}

	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	return nil
}
