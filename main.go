package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
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

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	username := os.Getenv("PGUSER")
	password := os.Getenv("PGPASSWORD")
	host := os.Getenv("PGHOST")

	db, err := internal.Connect(username, password, host)
	if err != nil {
		log.Fatalf("Error when connecting to the database: %s", err.Error())
	}

	stmt, err := db.Prepare("INSERT INTO dedupe.file_entry (scan_id, name, size, mode, mod_time, is_dir, hash) VALUES ($1, $2, $3, $4, $5, $6, $7)")
	if err != nil {
		log.Fatalf("Error when preparing statement: %s", err.Error())
	}

	scanId, err := uuid.NewRandom()
	if err != nil {
		log.Fatalf("Error when scan id: %s", err.Error())
	}

	gitignore := gitignore.CompileIgnoreLines(".git", "node_modules", ".yarn", ".git", ".tox", ".venv/", "target/", "build/", "dist/", "*.pyc")

	filenames, err := internal.GetFileNames(gitignore, root)
	if err != nil {
		panic("Error when fetching files: " + err.Error())
	}

	numWorkers := 100
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
