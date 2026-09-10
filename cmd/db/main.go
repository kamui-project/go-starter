package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

const migration = `
CREATE TABLE IF NOT EXISTS greetings (
    id SERIAL PRIMARY KEY,
    message VARCHAR(200) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Imported tables may rely on an ORM to supply the timestamp.
-- Go inserts only the message, so existing tables need this default too.
ALTER TABLE greetings ALTER COLUMN created_at SET DEFAULT NOW();
`

func main() {
	dsn := parseDatabaseURL(os.Getenv("DATABASE_URL"))
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	_, err = db.Exec(migration)
	if err != nil {
		panic(err)
	}
	fmt.Println("Migration complete")
}

func parseDatabaseURL(raw string) string {
	raw = strings.TrimPrefix(raw, "postgresql://")
	raw = strings.TrimPrefix(raw, "postgres://")

	userInfo, rest, _ := strings.Cut(raw, "@")
	user, password, _ := strings.Cut(userInfo, ":")
	hostPort, dbname, _ := strings.Cut(rest, "/")
	host, port, _ := strings.Cut(hostPort, ":")

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		host, port, user, password, dbname,
	)
}
