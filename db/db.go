package db

import (
	"database/sql"
	"log"
	"os"

	"github.com/cyb3rplis/discord-bot-go/config"

	_ "github.com/mattn/go-sqlite3"
)

var Config = config.GetConfig()

// InitDB initializes the SQLite database and loads the schema.
func InitDB() {
	databaseFile := Config.DB
	schemaFile := Config.Schema

	// Check if the database file exists
	if _, err := os.Stat(databaseFile); os.IsNotExist(err) {
		log.Println("Database does not exist, creating a new one...")
		file, err := os.Create(databaseFile)
		if err != nil {
			log.Fatalf("Failed to create database file: %v", err)
			return
		}
		err = file.Close()
		if err != nil {
			log.Fatalf("Failed to close database file: %v", err)
			return
		}
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
		err = tx.Rollback() // Rollback in case of an error
		if err != nil {
			log.Fatalf("Failed to rollback transaction: %v", err)
		}
		log.Fatalf("Failed to execute schema: %v", err)
	}

	if err = tx.Commit(); err != nil {
		log.Fatal(err)
	}

	log.Println("Database schema initialized successfully!")
}
