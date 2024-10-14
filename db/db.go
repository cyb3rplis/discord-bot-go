package db

import (
	"database/sql"
	"log"
	"os"

	"github.com/cyb3rplis/discord-bot-go/config"

	_ "github.com/mattn/go-sqlite3"
)

// InitDB initializes the SQLite database and loads the schema.
func InitDB() {
	cfg := config.GetConfig()
	databaseFile := cfg.DB
	schemaFile := cfg.Schema

	// Check if the database file exists
	if _, err := os.Stat(databaseFile); os.IsNotExist(err) {
		log.Println("Database does not exist, creating a new one...")
		file, err := os.Create(databaseFile)
		if err != nil {
			log.Fatalf("Failed to create database file: %v", err)
			return
		}
		file.Close()
		log.Println("Database created successfully!")
	} else {
		log.Println("Database already exists.")
	}

	// Open the SQLite database
	db, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Read the schema file
	schema, err := os.ReadFile(schemaFile)
	if err != nil {
		log.Fatalf("Failed to read schema file: %v", err)
		return
	}

	// Execute the schema file contents within a transaction
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	_, err = tx.Exec(string(schema))
	if err != nil {
		tx.Rollback() // Rollback in case of an error
		log.Fatalf("Failed to execute schema: %v", err)
	}

	if err = tx.Commit(); err != nil {
		log.Fatal(err)
	}

	log.Println("Database schema initialized successfully!")
}
