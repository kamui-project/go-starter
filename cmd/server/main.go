package main

import (
	"database/sql"
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

//go:embed templates/*
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

type Greeting struct {
	ID        int
	Message   string
	CreatedAt time.Time
}

func main() {
	db := connectDB()
	defer db.Close()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	tmpl := template.Must(
		template.New("").Funcs(template.FuncMap{
			"formatTime": func(t time.Time) string {
				return t.Format("2006-01-02 15:04")
			},
		}).ParseFS(templateFS, "templates/*"),
	)
	r.SetHTMLTemplate(tmpl)

	r.GET("/static/*filepath", func(c *gin.Context) {
		p := c.Param("filepath")
		data, err := staticFS.ReadFile("static" + p)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		http.ServeContent(c.Writer, c.Request, p, time.Time{}, bytes.NewReader(data))
	})

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.GET("/greetings", func(c *gin.Context) {
		rows, err := db.Query("SELECT id, message, created_at FROM greetings ORDER BY created_at DESC")
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		defer rows.Close()

		var greetings []Greeting
		for rows.Next() {
			var g Greeting
			rows.Scan(&g.ID, &g.Message, &g.CreatedAt)
			greetings = append(greetings, g)
		}
		c.HTML(http.StatusOK, "greetings.html", gin.H{"greetings": greetings})
	})

	r.POST("/greetings", func(c *gin.Context) {
		message := c.PostForm("message")
		if message != "" {
			db.Exec("INSERT INTO greetings (message) VALUES ($1)", message)
		}
		c.Redirect(http.StatusSeeOther, "/greetings")
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	r.Run("0.0.0.0:" + port)
}

func connectDB() *sql.DB {
	dsn := parseDatabaseURL(os.Getenv("DATABASE_URL"))
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}
	return db
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
