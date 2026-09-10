package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/lib/pq"
)

func TestMigrationPostgres(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to a dedicated PostgreSQL test database")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	for _, imported := range []bool{false, true} {
		name := "new table"
		if imported {
			name = "imported table without timestamp default"
		}
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback()

			schema := pq.QuoteIdentifier(fmt.Sprintf("greetings_migration_test_%d", time.Now().UnixNano()))
			exec := func(query string) {
				t.Helper()
				if _, err := tx.ExecContext(ctx, query); err != nil {
					t.Fatal(err)
				}
			}
			exec("CREATE SCHEMA " + schema)
			exec("SET LOCAL search_path TO " + schema)

			originalTime := time.Date(2026, 3, 29, 15, 9, 1, 0, time.UTC)
			if imported {
				exec(`CREATE TABLE greetings (
					id SERIAL PRIMARY KEY,
					message VARCHAR(200) NOT NULL,
					created_at TIMESTAMP NOT NULL
				)`)
				if _, err := tx.ExecContext(ctx,
					"INSERT INTO greetings (message, created_at) VALUES ($1, $2)",
					"existing greeting", originalTime); err != nil {
					t.Fatal(err)
				}
			}

			// Each deployment reruns the migration. Both runs must allow an insert
			// that omits created_at, just like the application's POST handler.
			for run := 1; run <= 2; run++ {
				exec(migration)
				var createdAt time.Time
				err := tx.QueryRowContext(ctx,
					"INSERT INTO greetings (message) VALUES ($1) RETURNING created_at",
					fmt.Sprintf("greeting after migration %d", run)).Scan(&createdAt)
				if err != nil {
					t.Fatalf("insert after migration %d: %v", run, err)
				}
				if createdAt.IsZero() {
					t.Fatalf("insert after migration %d has no timestamp", run)
				}
			}

			wantCount := 2
			if imported {
				wantCount++
				var message string
				var createdAt time.Time
				if err := tx.QueryRowContext(ctx,
					"SELECT message, created_at FROM greetings WHERE id = 1").
					Scan(&message, &createdAt); err != nil {
					t.Fatal(err)
				}
				if message != "existing greeting" || !createdAt.Equal(originalTime) {
					t.Fatalf("existing row changed: message=%q, created_at=%v", message, createdAt)
				}
			}
			var count int
			if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM greetings").Scan(&count); err != nil {
				t.Fatal(err)
			}
			if count != wantCount {
				t.Fatalf("row count = %d, want %d", count, wantCount)
			}
		})
	}
}
