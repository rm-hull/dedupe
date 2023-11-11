package internal

import (
	"database/sql"
	"fmt"
	"path/filepath"

	pg "dedupe/internal/db"
	"dedupe/internal/files"

	"github.com/gammazero/workerpool"
	"github.com/schollz/progressbar/v3"

	gitignore "github.com/sabhiram/go-gitignore"
)

func Scan(db *sql.DB, path string) error {

	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("unable to obtain root path: %w", err)
	}

	scanId, err := pg.CreateScan(db, absolutePath)
	if err != nil {
		return fmt.Errorf("error when creating scan: %w", err)
	}

	stmt, err := pg.InsertFileEntryStatement(db)
	if err != nil {
		return fmt.Errorf("error when preparing insert statement: %w", err)
	}

	gitignore := gitignore.CompileIgnoreLines(".git", "node_modules", ".yarn", ".tox", ".venv/", ".ivy", "target/", "build/", "dist/", "*.pyc", "*.jar")
	bar1 := progressbar.Default(-1, "[1/2] Counting files")

	filenames, err := files.GetFileNames(gitignore, absolutePath, func() { bar1.Add(1) })
	if err != nil {
		return fmt.Errorf("error when fetching files: %w", err)
	}

	numWorkers := 100 // 100 is good for macOS, not so good for Linux
	wp := workerpool.New(numWorkers)

	numFiles := len(filenames)
	bar2 := progressbar.Default(int64(numFiles), "[2/2] Indexing files")

	for _, filename := range filenames {
		localFilename := filename
		wp.Submit(func() {
			defer bar2.Add(1)

			file, err := files.GetFileDetails(localFilename)

			if err != nil {
				// TODO: log the error to the db
				fmt.Println(err.Error())
			} else {
				_, err = stmt.Exec(scanId, file.Name, file.Size, file.Mode, file.ModTime, file.IsDir, file.Hash)
				if err != nil {
					panic(err) // FIXME: should be handled properly
				}
			}
		})
	}

	wp.StopWait()
	return nil
}
