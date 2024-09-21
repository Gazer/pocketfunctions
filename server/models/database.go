package models

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB() *sql.DB {
	db, _ := sql.Open("sqlite3", "./db.sqlite3")
	statement, err := db.Prepare(`CREATE TABLE IF NOT EXISTS functions (
		id INTEGER PRIMARY KEY,
		uri VARCHAR(255),
		code VARCHAR(255)
	)`)
	if err != nil {
		log.Println("Error in creating table")
	} else {
		log.Println("Successfully created table books!")
	}
	statement.Exec()
	return db
}

func CreateFunction(db *sql.DB, uri string) (int64, error) {
	function, err := GetFunctionByUri(db, uri)
	if err == nil {
		log.Printf("Function exists with id=%d\n", function.Id)
		return int64(function.Id), nil
	}

	statement, _ := db.Prepare("INSERT INTO functions (uri, code) VALUES (?, ?)")
	result, err := statement.Exec(uri, "")
	if err != nil {
		log.Println(err.Error())
		return -1, err
	}
	return result.LastInsertId()
}

func UpdateFunction(db *sql.DB, function *PocketFunction) error {
	statement, _ := db.Prepare("UPDATE functions SET uri=?, code=? WHERE id=?")
	_, err := statement.Exec(function.Uri, function.Code, function.Id)
	return err
}

func GetFunctionByID(db *sql.DB, id string) (*PocketFunction, error) {
	statement, _ := db.Prepare("SELECT * FROM functions WHERE id=?")
	rows, _ := statement.Query(id)
	defer rows.Close()

	var function PocketFunction
	rows.Next()
	rows.Scan(&function.Id, &function.Uri, &function.Code)

	return &function, nil
}

func GetFunctionByUri(db *sql.DB, uri string) (*PocketFunction, error) {
	statement, _ := db.Prepare("SELECT * FROM functions WHERE uri=?")
	rows, _ := statement.Query(uri)
	defer rows.Close()

	var function PocketFunction
	if rows.Next() {
		rows.Scan(&function.Id, &function.Uri, &function.Code)

		return &function, nil
	}

	return nil, fmt.Errorf("Not found")
}