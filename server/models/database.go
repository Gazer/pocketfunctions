package models

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Pair[T, U any] struct {
	First  T
	Second U
}

func InitDB() *sql.DB {
	db, _ := sql.Open("sqlite3", "./db.sqlite3")
	createFunctionsTable(db)
	createFunctionsAnalyticsTable(db)
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

func RegisterExecuted(db *sql.DB, id int, elapsed time.Duration, status int16) {
	statement, _ := db.Prepare(`
		INSERT INTO functions_analytics (function_id, time_ms, status)
		VALUES (?, ?, ?)
	`)
	_, err := statement.Exec(id, elapsed.Milliseconds(), status)
	if err != nil {
		log.Println(err.Error())
	}
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

func GetFunctions(db *sql.DB) ([]*PocketFunction, error) {
	rows, err := db.Query(`
		SELECT functions.*, count(function_id) as executions, AVG(time_ms) as average
		FROM functions
		INNER JOIN functions_analytics on
			functions_analytics.function_id = functions.id
		GROUP BY functions.id;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var functions []*PocketFunction

	for rows.Next() {
		var function PocketFunction
		rows.Scan(&function.Id, &function.Uri, &function.Code, &function.Execution, &function.Average)
		functions = append(functions, &function)
	}

	return functions, nil
}

func GetHistogram(db *sql.DB) ([]Pair[string, int], error) {
	rows, err := db.Query(`WITH RECURSIVE date_series AS (
	    SELECT date('now', '-29 days') AS date
	    UNION ALL
	    SELECT date(date, '+1 day')
	    FROM date_series
	    WHERE date < date('now')
		)
		SELECT ds.date AS day, COUNT(fa.id) AS count
		FROM date_series ds
		LEFT JOIN functions_analytics fa
		ON date(fa.created_at) = ds.date
		GROUP BY ds.date
		ORDER BY ds.date;
	`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var pairs []Pair[string, int]

	for rows.Next() {
		var pair Pair[string, int]
		rows.Scan(&pair.First, &pair.Second)
		pairs = append(pairs, pair)
	}

	return pairs, nil
}

func GetTotalCalls(db *sql.DB) (Pair[int, float32], error) {
	rows, err := db.Query(`WITH current_week AS (
	    SELECT COUNT(*) AS count
	    FROM functions_analytics
	    WHERE created_at >= date('now', 'weekday 0', '-6 days')
	    AND created_at < date('now', 'weekday 0', '+1 day')
			AND status = 200
    ),
    previous_week AS (
	    SELECT COUNT(*) AS count
	    FROM functions_analytics
	    WHERE created_at >= date('now', 'weekday 0', '-13 days')
	    AND created_at < date('now', 'weekday 0', '-6 days')
			AND status = 200
		)
		SELECT cw.count AS current_week_count,
    CASE
        WHEN pw.count = 0 AND cw.count > 0 THEN 100.0
        WHEN pw.count = 0 AND cw.count = 0 THEN 0.0
        ELSE ROUND(((cw.count - pw.count) * 100.0) / pw.count, 2)
    END AS percentage_variation
    FROM current_week cw, previous_week pw;
  `)

	var pair Pair[int, float32]

	if err != nil {
		return pair, err
	}

	defer rows.Close()

	if rows.Next() {
		err := rows.Scan(&pair.First, &pair.Second)
		return pair, err
	}

	return pair, fmt.Errorf("No rows")
}

func GetTotalErrors(db *sql.DB) (Pair[int, float32], error) {
	rows, err := db.Query(`WITH current_week AS (
	    SELECT COUNT(*) AS count
	    FROM functions_analytics
	    WHERE created_at >= date('now', 'weekday 0', '-6 days')
	    AND created_at < date('now', 'weekday 0', '+1 day')
			AND status <> 200
    ),
    previous_week AS (
	    SELECT COUNT(*) AS count
	    FROM functions_analytics
	    WHERE created_at >= date('now', 'weekday 0', '-13 days')
	    AND created_at < date('now', 'weekday 0', '-6 days')
			AND status <> 200
		)
		SELECT cw.count AS current_week_count,
    CASE
        WHEN pw.count = 0 AND cw.count > 0 THEN 100.0
        WHEN pw.count = 0 AND cw.count = 0 THEN 0.0
        ELSE ROUND(((cw.count - pw.count) * 100.0) / pw.count, 2)
    END AS percentage_variation
    FROM current_week cw, previous_week pw;
  `)

	var pair Pair[int, float32]

	if err != nil {
		return pair, err
	}

	defer rows.Close()

	if rows.Next() {
		err := rows.Scan(&pair.First, &pair.Second)
		return pair, err
	}

	return pair, fmt.Errorf("No rows")
}

func GetAvgTime(db *sql.DB) (Pair[int, float32], error) {
	rows, err := db.Query(`WITH current_week AS (
	    SELECT AVG(time_ms) AS count
	    FROM functions_analytics
	    WHERE created_at >= date('now', 'weekday 0', '-6 days')
	    AND created_at < date('now', 'weekday 0', '+1 day')
			AND status = 200
		),
		previous_week AS (
			SELECT AVG(time_ms) AS count
			FROM functions_analytics
			WHERE created_at >= date('now', 'weekday 0', '-13 days')
			AND created_at < date('now', 'weekday 0', '-6 days')
			AND status = 200
		)
		SELECT
    cw.count AS current_week_count,
    CASE
        WHEN pw.count = 0 AND cw.count > 0 THEN 100.0
        WHEN pw.count = 0 AND cw.count = 0 THEN 0.0
        ELSE ROUND(((cw.count - pw.count) * 100.0) / pw.count, 2)
    END AS percentage_variation
    FROM current_week cw, previous_week pw;
  `)
	var pair Pair[int, float32]

	if err != nil {
		return pair, err
	}

	defer rows.Close()

	if rows.Next() {
		err := rows.Scan(&pair.First, &pair.Second)
		return pair, err
	}

	return pair, fmt.Errorf("No rows")
}

func createFunctionsTable(db *sql.DB) {
	statement, err := db.Prepare(`CREATE TABLE IF NOT EXISTS functions (
		id INTEGER PRIMARY KEY,
		uri VARCHAR(255),
		code VARCHAR(255)
	)`)
	if err != nil {
		log.Println("Error in creating table")
	} else {
		log.Println("Successfully created table functions!")
	}
	statement.Exec()
}

func createFunctionsAnalyticsTable(db *sql.DB) {
	statement, err := db.Prepare(`CREATE TABLE IF NOT EXISTS functions_analytics (
		id INTEGER PRIMARY KEY,
		function_id INTEGER NOT NULL,
		time_ms INTEGER NOT NULL,
		status INTEGER NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		log.Println("Error in creating table")
	} else {
		log.Println("Successfully created table functions_analytics!")
	}
	statement.Exec()
}
