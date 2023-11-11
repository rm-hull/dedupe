package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"runtime"

	"dedupe/internal"
	pg "dedupe/internal/db"

	"github.com/carlmjohnson/versioninfo"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
)

func main() {
	db, err := initDatabase()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	app := &cli.App{
		Name:                 "dedupe",
		Version:              versioninfo.Short(),
		Copyright:            "(c) 2023 Richard Hull",
		Usage:                "Scans and identifies duplicate files across machines and file-systems",
		UsageText:            "...",
		Suggest:              true,
		EnableBashCompletion: true,
		Authors: []*cli.Author{
			{
				Name:  "Richard Hull",
				Email: "rm_hull@yahoo.co.uk",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "scan",
				Usage: "scans a filesystem path",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "path",
						Aliases:  []string{"p"},
						Required: true,
						Usage:    "The file-system path to scan (can be absolute or relative)",
					},
					&cli.IntFlag{
						Name:    "num-workers",
						Aliases: []string{"n"},
						Value:   getGoodWorkerCount(),
						Usage:   "The number of workers to spit scanning into: 100 is good for macOS with a fast SSD, not so good for Linux with a spinnging HDD, where a value of 10 might be more appropriate",
					},
				},
				Action: func(cCtx *cli.Context) error {
					fmt.Println("scan: ", cCtx.String("path"))
					path := cCtx.String("path")
					numWorkers := cCtx.Int("num-workers")
					return internal.Scan(db, path, numWorkers)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func initDatabase() (*sql.DB, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	username := os.Getenv("PGUSER")
	password := os.Getenv("PGPASSWORD")
	host := os.Getenv("PGHOST")

	db, err := pg.Connect(username, password, host)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to the database: %w", err)
	}

	err = pg.Migrate(db)
	if err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return db, nil
}

func getGoodWorkerCount() int {
	switch os := runtime.GOOS; os {
	case "darwin":
		return 100
	default:
		return 10
	}
}
