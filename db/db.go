package db

import (
	"database/sql"
	"github.com/cyb3rplis/discord-bot-go/model"
	"log"
	"os"

	"github.com/cyb3rplis/discord-bot-go/config"

	_ "github.com/mattn/go-sqlite3"
)

var Config = config.GetConfig()

// InitModel initializes the SQLite database schema
func InitModel() (model.Model, func() error, error) {
	db, dbClose, err := InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
		return model.Model{}, nil, err
	}
	m := model.Model{
		Db: db,
	}
	return m, dbClose, nil
}

func InitDB() (*sql.DB, func() error, error) {
	databaseFile := Config.DB
	schemaFile := Config.Schema

	// Check if the database file exists
	if _, err := os.Stat(databaseFile); os.IsNotExist(err) {
		log.Println("Database does not exist, creating a new one...")
		file, err := os.Create(databaseFile)
		if err != nil {
			log.Fatalf("Failed to create database file: %v", err)
			return nil, nil, err
		}
		err = file.Close()
		if err != nil {
			log.Fatalf("Failed to close database file: %v", err)
			return nil, nil, err
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

	// Read the schema file
	schema, err := os.ReadFile(schemaFile)
	if err != nil {
		log.Fatalf("Failed to read schema file: %v", err)
		return nil, nil, err
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

	return db, db.Close, nil
}
