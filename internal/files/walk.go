package files

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"time"

	_ "github.com/lib/pq"

	gitignore "github.com/sabhiram/go-gitignore"
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

func GetFileNames(gitignore *gitignore.GitIgnore, root string, callback func() error) ([]string, error) {
	files := make([]string, 0, 2000)
	visit := func(path string, dirEntry fs.DirEntry, err error) error {
		defer func() {
			err = callback()
		}()

		if !gitignore.MatchesPath(path) {
			files = append(files, path)
		}
		return err
	}

	err := filepath.WalkDir(root, visit)
	return files, err
}
