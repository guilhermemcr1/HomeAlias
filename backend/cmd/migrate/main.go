package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	dir := flag.String("path", "file://migrations", "migrations path")
	flag.Parse()
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		log.Fatal("DATABASE_DSN required")
	}
	// migrate mysql driver expects mysql://user:pass@tcp(host:port)/db
	m, err := migrate.New(*dir, "mysql://"+dsn)
	if err != nil {
		log.Fatalf("migrate new: %v", err)
	}
	defer m.Close()
	cmd := "up"
	if flag.NArg() > 0 {
		cmd = flag.Arg(0)
	}
	switch cmd {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("up: %v", err)
		}
		fmt.Println("migrations applied")
	case "down":
		if err := m.Steps(-1); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("down: %v", err)
		}
		fmt.Println("migrated down one step")
	default:
		log.Fatalf("unknown command %s", cmd)
	}
}
