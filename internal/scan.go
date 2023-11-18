package internal

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"

	pg "dedupe/internal/db"
	"dedupe/internal/files"

	"github.com/gammazero/workerpool"
	"github.com/schollz/progressbar/v3"

	gitignore "github.com/sabhiram/go-gitignore"
)

var commonIgnores = []string{
	// Javascript
	"node_modules", ".yarn",
	// Python
	"*.pyc", ".tox", ".venv/",
	// Java/Scala/Kotlin etc
	"*.jar", "*.class", ".ivy", ".m2", ".sbt",
	// Misc
	".git", "target/", "build/", "dist/",
}

func Scan(db *sql.DB, path string, numWorkers int, ignorepath string) error {

	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("unable to obtain root path: %w", err)
	}

	stmt, err := pg.InsertFileEntryStatement(db)
	if err != nil {
		return fmt.Errorf("error when preparing insert statement: %w", err)
	}

	var ignore *gitignore.GitIgnore
	if ignorepath != "" {
		ignore, err = gitignore.CompileIgnoreFileAndLines(ignorepath, commonIgnores...)
		if err != nil {
			return fmt.Errorf("error when compile ignore patterns: %w", err)
		}
	} else {
		ignore = gitignore.CompileIgnoreLines(commonIgnores...)
	}

	bar1 := progressbar.Default(-1, "[1/2] Counting files")

	filenames, err := files.GetFileNames(ignore, absolutePath, func() error { return bar1.Add(1) })
	if err != nil {
		return fmt.Errorf("error when fetching files: %w", err)
	}
	numFiles := len(filenames)
	pool := workerpool.New(numWorkers)
	bar2 := progressbar.Default(int64(numFiles), "[2/2] Indexing files")

	scanId, err := pg.CreateScan(db, absolutePath)
	if err != nil {
		return fmt.Errorf("error when creating scan: %w", err)
	}

	defer func() {
		if err := pg.UpdateScan(db, *scanId, numFiles, err); err != nil {
			log.Fatal(err)
		}
	}()

	for _, filename := range filenames {
		localFilename := filename
		pool.Submit(func() {
			defer func() {
				if err := bar2.Add(1); err != nil {
					log.Fatal(err)
				}
			}()

			file, err := files.GetFileDetails(localFilename)

			if err != nil {
				// TODO: log the error to the db
				fmt.Println(err.Error())
			} else {
				_, err = stmt.Exec(scanId, file.Name, file.Size, file.Mode, file.ModTime, &file.Hash)
				if err != nil {
					log.Fatal(err) // FIXME: should be handled properly
				}
			}
		})
	}

	pool.StopWait()
	return nil
}
