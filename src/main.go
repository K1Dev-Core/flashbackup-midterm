package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	dbPath := flag.String("db", "crossplatform.db", "SQLite database path")
	flag.Parse()

	db, err := gorm.Open(sqlite.Open(*dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		fatal(err)
	}
	if err := db.AutoMigrate(&File{}, &Setting{}); err != nil {
		fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		fatal(err)
	}
	defer sqlDB.Close()

	app := &App{db: db}
	app.loadSettings()
	app.printHeader()
	app.run()
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
