package internal

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	_ "github.com/lib/pq"

	gitignore "github.com/sabhiram/go-gitignore"

	"github.com/schollz/progressbar/v3"
)

type File struct {
	Name    string
	Size    int64
	Mode    fs.FileMode
	ModTime time.Time
	IsDir   bool
	Hash    string
}

func (f *File) ToString() string {
	return fmt.Sprintf("%s (isDir=%t, size=%d, hash=%s)", f.Name, f.IsDir, f.Size, f.Hash)
}

func GetFileNames(gitignore *gitignore.GitIgnore, root string) ([]string, error) {

	files := make([]string, 0)
	bar := progressbar.Default(-1, "[1/2] Counting files")

	visit := func(path string, dirEntry fs.DirEntry, err error) error {
		bar.Add(1)
		if !gitignore.MatchesPath(path) {
			files = append(files, path)
		}
		return err
	}

	err := filepath.WalkDir(root, visit)
	return files, err
}

func GetFileDetails(filename string) (*File, error) {
	fi, err := os.Lstat(filename)
	if err != nil {
		return nil, err
	}

	hash, _ := Hash(filename)

	file := &File{
		Name:    filename,
		Size:    fi.Size(),
		Mode:    fi.Mode(),
		ModTime: fi.ModTime(),
		IsDir:   fi.IsDir(),
		Hash:    hash,
	}

	return file, nil
}

func Hash(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
