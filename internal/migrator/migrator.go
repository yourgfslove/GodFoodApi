package migrator

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func MustMigrate(storagePath, migrationsPath string) {
	db, err := sql.Open("postgres", storagePath)
	if err != nil {
		panic(fmt.Sprintf("failed to open database: %v", err))
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		panic(fmt.Sprintf("failed to set dialect: %v", err))
	}
	if err := goose.Up(db, migrationsPath); err != nil {
		panic(fmt.Sprintf("failed to migrate: %v", err))
	}
}
