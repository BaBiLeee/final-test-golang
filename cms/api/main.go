package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	elastic "github.com/olivere/elastic/v7"

	"yourmodule/internal/cache"
	"yourmodule/internal/db"
	"yourmodule/internal/es"
	"yourmodule/internal/handlers"

	"github.com/gorilla/mux"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	redisAddr := os.Getenv("REDIS_ADDR")
	esURL := os.Getenv("ES_URL")

	// Postgres
	sqlDB, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	sqlDB.SetConnMaxLifetime(time.Minute * 3)
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)

	// Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// Elasticsearch client
	esClient, err := elastic.NewClient(elastic.SetURL(esURL), elastic.SetSniff(false))
	if err != nil {
		log.Fatal(err)
	}

	// initialize lower-level modules
	dbRepo := db.New(sqlDB)
	cacheClient := cache.New(rdb)
	esService := es.New(esClient)

	// Handlers
	h := handlers.New(dbRepo, cacheClient, esService)

	r := mux.NewRouter()
	r.HandleFunc("/posts", h.CreatePost).Methods("POST")
	r.HandleFunc("/posts/{id:[0-9]+}", h.GetPost).Methods("GET")
	r.HandleFunc("/posts/{id:[0-9]+}", h.UpdatePost).Methods("PUT")
	r.HandleFunc("/posts/search-by-tag", h.SearchByTag).Methods("GET")
	r.HandleFunc("/posts/search", h.SearchES).Methods("GET")

	srv := &http.Server{
		Handler:      r,
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("Listening on :8080")
	log.Fatal(srv.ListenAndServe())
}
