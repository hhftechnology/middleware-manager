//go:build !cgo

package database

import (
	"database/sql"

	"modernc.org/sqlite"
)

func init() {
	sql.Register("sqlite3", &sqlite.Driver{})
}
