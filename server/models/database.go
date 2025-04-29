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

func CreateFunction(db *sql.DB, name string, lang string) (int64, error) {
	function, err := GetFunctionByName(db, name)
	if err == nil {
		log.Printf("Function exists with id=%d\n", function.Id)
		return int64(function.Id), nil
	}

	statement, err := db.Prepare("INSERT INTO functions (name, docker_id, lang) VALUES (?, ?, ?)")
	if err != nil {
		return -1, err
	}
	result, err := statement.Exec(name, "", lang)
	if err != nil {
		log.Println(err.Error())
		return -1, err
	}
	return result.LastInsertId()
}

func UpdateFunction(db *sql.DB, function *PocketFunction) error {
	statement, err := db.Prepare("UPDATE functions SET name=?, docker_id=?, lang=? WHERE id=?")
	if err != nil {
		return err
	}
	_, err = statement.Exec(function.Name, function.DockerId, function.Lang, function.Id)
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
	rows.Scan(&function.Id, &function.Name, &function.Lang, &function.DockerId)

	return &function, nil
}

func GetFunctionByName(db *sql.DB, name string) (*PocketFunction, error) {
	statement, _ := db.Prepare("SELECT * FROM functions WHERE name=?")
	rows, _ := statement.Query(name)
	defer rows.Close()

	var function PocketFunction
	if rows.Next() {
		rows.Scan(&function.Id, &function.Name, &function.Lang, &function.DockerId)

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
		rows.Scan(&function.Id, &function.Name, &function.DockerId, &function.Name, &function.Execution, &function.Average)
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
	    SELECT COUNT(*) AS value
	    FROM functions_analytics
	    WHERE created_at >= date('now', 'weekday 0', '-6 days')
	    AND created_at < date('now', 'weekday 0', '+1 day')
			AND status = 200
    ),
    previous_week AS (
	    SELECT COUNT(*) AS value
	    FROM functions_analytics
	    WHERE created_at >= date('now', 'weekday 0', '-13 days')
	    AND created_at < date('now', 'weekday 0', '-6 days')
			AND status = 200
		)
		SELECT cw.value AS current_week_value,
    CASE
        WHEN pw.value = 0 AND cw.value > 0 THEN 100.0
        WHEN pw.value = 0 AND cw.value = 0 THEN 0.0
        ELSE ROUND(((cw.value - pw.value) * 100.0) / pw.value, 2)
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
	    SELECT COUNT(*) AS value
	    FROM functions_analytics
	    WHERE created_at >= date('now', 'weekday 0', '-6 days')
	    AND created_at < date('now', 'weekday 0', '+1 day')
			AND status <> 200
    ),
    previous_week AS (
	    SELECT COUNT(*) AS value
	    FROM functions_analytics
	    WHERE created_at >= date('now', 'weekday 0', '-13 days')
	    AND created_at < date('now', 'weekday 0', '-6 days')
			AND status <> 200
		)
		SELECT cw.value AS current_week_value,
    CASE
        WHEN pw.value = 0 AND cw.value > 0 THEN 100.0
        WHEN pw.value = 0 AND cw.value = 0 THEN 0.0
        ELSE ROUND(((cw.value - pw.value) * 100.0) / pw.value, 2)
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

func GetAvgTime(db *sql.DB) (Pair[float64, float64], error) {
	rows, err := db.Query(`WITH current_week AS (
	    SELECT IFNULL(AVG(time_ms),0) AS value
			FROM functions_analytics
	    WHERE created_at >= date('now', 'weekday 0', '-6 days')
	    AND created_at < date('now', 'weekday 0', '+1 day')
			AND status = 200
		),
		previous_week AS (
			SELECT IFNULL(AVG(time_ms), 0) AS value
			FROM functions_analytics
			WHERE created_at >= date('now', 'weekday 0', '-13 days')
			AND created_at < date('now', 'weekday 0', '-6 days')
			AND status = 200
		)
		SELECT
    cw.value AS current_week_value,
    CASE
        WHEN pw.value = 0 AND cw.value > 0 THEN 100.0
        WHEN pw.value = 0 AND cw.value = 0 THEN 0.0
        ELSE ROUND(((cw.value - pw.value) * 100.0) / pw.value, 2)
    END AS percentage_variation
    FROM current_week cw, previous_week pw;
  `)
	var pair Pair[float64, float64]

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
		name VARCHAR(255),
		lang VARCHAR(255),
		docker_id VARCHAR(255)
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
