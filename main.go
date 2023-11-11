package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"rm-hull/dedupe/internal"

	"github.com/gammazero/workerpool"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	gitignore "github.com/sabhiram/go-gitignore"
	"github.com/schollz/progressbar/v3"
)

func WriteToDB(bar *progressbar.ProgressBar, scanId uuid.UUID, stmt *sql.Stmt, filename string, ch chan<- *internal.File) {

}

func main() {
	flag.Parse()
	root := flag.Arg(0)
	absolutePath, err := filepath.Abs(root)
	if err != nil {
		log.Fatalf("Unable to obtain root path: %s", err.Error())
	}

	err = godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err.Error())
	}

	username := os.Getenv("PGUSER")
	password := os.Getenv("PGPASSWORD")
	host := os.Getenv("PGHOST")

	db, err := internal.Connect(username, password, host)
	if err != nil {
		log.Fatalf("Error when connecting to the database: %s", err.Error())
	}

	err = internal.Migrate(db)
	if err != nil {
		log.Fatalf("Error when migrating the database: %s", err.Error())
	}

	scanId, err := internal.CreateScan(db, absolutePath)
	if err != nil {
		log.Fatalf("Error when creating scan: %s", err.Error())
	}

	stmt, err := internal.InsertFileEntryStatement(db)
	if err != nil {
		log.Fatalf("Error when preparing statement: %s", err.Error())
	}

	gitignore := gitignore.CompileIgnoreLines(".git", "node_modules", ".yarn", ".tox", ".venv/", ".ivy", "target/", "build/", "dist/", "*.pyc", "*.jar")

	filenames, err := internal.GetFileNames(gitignore, absolutePath)
	if err != nil {
		panic("Error when fetching files: " + err.Error())
	}

	numWorkers := 100 // 100 is good for macOS, not so good for Linux
	wp := workerpool.New(numWorkers)

	numFiles := len(filenames)
	bar := progressbar.Default(int64(numFiles), "[2/2] Indexing files")

	defer db.Close()

	for _, filename := range filenames {
		localFilename := filename
		wp.Submit(func() {
			defer bar.Add(1)

			file, err := internal.GetFileDetails(localFilename)

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
	bar.RenderBlank()
}
