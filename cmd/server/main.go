package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/xoltia/setlink/internal/controllers"
	"github.com/xoltia/setlink/internal/models"
	"github.com/xoltia/setlink/internal/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	Port   = flag.Int("port", 8080, "Port to listen on")
	DbPath = flag.String("db-path", "test.db", "Path to database file")
	//CrawlTimeout = flag.Duration("crawl-timeout", 10*time.Second, "Timeout for crawling a single URL")
	//RequestsPerSecond = flag.Int("requests-per-second", 2, "Maximum number of requests per second")
)

func main() {
	flag.Parse()

	log.Println("Starting server on port", *Port)

	db, err := gorm.Open(sqlite.Open(*DbPath), &gorm.Config{})
	linksetService := services.NewLinkSetService(db)
	apiController := controllers.NewAPIController(linksetService)
	staticController := controllers.NewStaticController(linksetService)

	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&models.LinkSet{})

	// rateLimit := NewRateLimiter(rate.Limit(*RequestsPerSecond), 1)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	//r.Use(rateLimit.GetLimiterMiddleware)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Mount("/api", apiController.Router())
	r.Mount("/", staticController.Router())

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *Port), r))
}
