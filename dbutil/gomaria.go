package dbutil

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

const (
	DataSourceName = "root:950918@tcp(192.168.0.110:3306)/robot?charset=utf8"
)

var Db *sql.DB

func init() {
	Db, _ = sql.Open("mysql", DataSourceName)
	Db.SetMaxOpenConns(2000)
	Db.SetMaxIdleConns(1000)
	Db.Ping()
}
