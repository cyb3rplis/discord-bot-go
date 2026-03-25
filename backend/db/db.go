package db

import (
	"database/sql"
	"embed"
	"os"
	"path/filepath"

	"github.com/cyb3rplis/discord-bot-go/config"
	log "github.com/cyb3rplis/discord-bot-go/logger"
	"github.com/cyb3rplis/discord-bot-go/model"

	_ "github.com/mattn/go-sqlite3"
)

var Config = config.GetConfig()

//go:embed schema.sql
var fs embed.FS

// InitModel initializes the SQLite database schema
func InitModel() (model.Model, func() error, error) {
	db, dbClose, err := InitDB()
	if err != nil {
		log.FatalLog.Printf("Failed to initialize database: %v", err)
		return model.Model{}, nil, err
	}
	m := model.Model{
		Db: db,
	}
	return m, dbClose, nil
}

func InitDB() (*sql.DB, func() error, error) {
	databaseFile := Config.DB

	// Check if the ../dist path exists
	if _, err := os.Stat(filepath.Join(config.AppPath(), "data")); os.IsNotExist(err) {
		log.FatalLog.Fatalln(filepath.Join(config.AppPath(), "data") + " directory does not exist, make sure it exists before starting the container!")
	}

	// Check if the database file exists
	if _, err := os.Stat(databaseFile); os.IsNotExist(err) {
		log.InfoLog.Printf("Database does not exist, creating a new one: %v", databaseFile)
		file, err := os.Create(databaseFile)
		if err != nil {
			log.FatalLog.Fatalf("Failed to create database file: %v", err)
			return nil, nil, err
		}
		err = file.Close()
		if err != nil {
			log.FatalLog.Fatalf("Failed to close database file: %v", err)
			return nil, nil, err
		}
		log.InfoLog.Printf("Database created successfully: %v", databaseFile)
	} else {
		log.InfoLog.Printf("Database already exists: %v", databaseFile)
	}

	// Open the SQLite database
	db, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		log.FatalLog.Fatal(err)
	}

	// Read the schema file
	schema, err := fs.ReadFile("schema.sql")
	if err != nil {
		log.FatalLog.Fatalf("Failed to read schema file: %v", err)
	}

	// Execute the schema file contents within a transaction
	tx, err := db.Begin()
	if err != nil {
		log.FatalLog.Fatal(err)
	}

	// Execute the schema file contents
	_, err = tx.Exec(string(schema))
	if err != nil {
		err = tx.Rollback()
		if err != nil {
			log.FatalLog.Fatalf("Failed to rollback transaction: %v", err)
		}
		log.FatalLog.Fatalf("Failed to execute schema: %v", err)
	}

	if err = tx.Commit(); err != nil {
		log.FatalLog.Fatal(err)
	}

	log.InfoLog.Println("Database schema initialized successfully")

	return db, db.Close, nil
}
