package util

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

func Get_DB(dbfile string) *sql.DB {
	db, err := sql.Open("sqlite3", dbfile)
	checkerr(err)
	return db
}

func Exists(db *sql.DB, cveid string) bool {
	rows, err := db.Query("select count(cveid) from status where cveid=?", cveid)
	checkerr(err)
	defer rows.Close()

	count := int(0)

	for rows.Next() {
		rows.Scan(&count)
	}

	if count > 0 {
		return true
	}

	// Apparently, no
	return false
}

func Modified_Matches(db *sql.DB, cveid string, modified time.Time) bool {
	rows, err := db.Query("select count(cveid) from status where cveid=? and modified=?", cveid, modified)
	checkerr(err)
	defer rows.Close()

	count := int(0)

	for rows.Next() {
		rows.Scan(&count)
	}

	if count > 0 {
		return true
	}

	// Apparently, no
	return false
}

func DB_Add(db *sql.DB, cveid string, modified time.Time, ticketid string) {
	_, err := db.Exec("insert into status(cveid, modified, ticketid) values (?, ?, ?)", cveid, modified, ticketid)
	checkerr(err)
}

func DB_Update(db *sql.DB, cveid string, modified time.Time) {
	_, err := db.Exec("update status set modified=? where cveid=?", modified, cveid)
	checkerr(err)
}

func DB_TicketID(db *sql.DB, cveid string) string {
	rows, err := db.Query("select ticketid from status where cveid=?", cveid)
	checkerr(err)
	defer rows.Close()

	id := ""

	for rows.Next() {
		rows.Scan(&id)
	}

	return id
}
