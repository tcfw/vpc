package vpc

import (
	"context"
	"database/sql"
	"io/ioutil"
	"log"
	"os"
	"time"

	//postgresql DB driver
	_ "github.com/lib/pq"
)

const (
	dbName = "vpc"
)

//DBConn opens a connection to a postgresql db server
//and attempts to validate the connection
func DBConn() (*sql.DB, error) {
	dbEP := os.Getenv("DB")
	if dbEP == "" {
		dbEP = "postgresql://localhost:26257"
	}

	db, err := sql.Open("postgres", dbEP)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

type migration struct {
	file       string
	migratedAt time.Time
}

//Migrate runs through each sql file in the migrations folder
//and attempts to execute it in the DB
func Migrate(db *sql.DB, dir string) error {
	statements := []string{
		`CREATE DATABASE IF NOT EXISTS ` + dbName,
		`CREATE TABLE IF NOT EXISTS ` + dbName + `.migrations (
			file STRING,
			migrated_at TIMESTAMP,
			PRIMARY KEY (file)
		)`,
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	migratedRows, err := db.Query("select file from migrations;")
	if err != nil {
		log.Fatal(err)
	}
	defer migratedRows.Close()

	migrations := map[string]bool{}
	if migratedRows != nil {
		for migratedRows.Next() {
			var file string
			if err := migratedRows.Scan(&file); err != nil {
				// Check for a scan error.
				// Query rows will be closed with defer.
				log.Fatal(err)
			}
			migrations[file] = true
		}
	}

	if len(files) == len(migrations) {
		log.Println("Nothing to migrate")
		return nil
	}

	log.Println("Starting migrations...")

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	for _, f := range files {
		if _, ok := migrations[f.Name()]; ok {
			continue
		}

		log.Printf("Running %s...\n", f.Name())

		migBytes, err := ioutil.ReadFile(dir + f.Name())
		if err != nil {
			tx.Rollback()
			return err
		}

		if _, err := tx.Exec(string(migBytes)); err != nil {
			tx.Rollback()
			return err
		}

		if _, err := tx.Exec(`insert into `+dbName+`.migrations VALUES ($1, NOW())`, f.Name()); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
