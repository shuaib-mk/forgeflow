package main

import (
	"context"
	"fmt"
	"os"

	"github.com/forgeflow/forgeflow/internal/database"
	"github.com/forgeflow/forgeflow/internal/migrations"
)

func main() {
	if len(os.Args) != 2 || os.Args[1] != "up" {
		fmt.Fprintln(os.Stderr, "usage: forgeflow-migrate up")
		os.Exit(2)
	}
	databaseURL := os.Getenv("FORGEFLOW_DATABASE_URL")
	if databaseURL == "" {
		fmt.Fprintln(os.Stderr, "FORGEFLOW_DATABASE_URL is required")
		os.Exit(2)
	}
	db, err := database.Open(context.Background(), databaseURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer db.Close()
	directory := os.Getenv("FORGEFLOW_MIGRATIONS_DIR")
	if directory == "" {
		directory = "deployments/migrations"
	}
	if err := migrations.Up(context.Background(), db.Pool, directory); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("database migrations applied")
}
