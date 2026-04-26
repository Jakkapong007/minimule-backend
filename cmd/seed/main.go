package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	sqlFile := "migrations/seed.sql"
	if len(os.Args) > 1 {
		sqlFile = os.Args[1]
	}
	sql, err := os.ReadFile(sqlFile)
	if err != nil {
		log.Fatalf("read %s: %v", sqlFile, err)
	}
	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer conn.Close(context.Background())

	_, err = conn.Exec(context.Background(), string(sql))
	if err != nil {
		log.Fatalf("exec seed: %v", err)
	}
	fmt.Println("seed complete")
}
