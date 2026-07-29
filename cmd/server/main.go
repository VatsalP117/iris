package main

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"

	"github.com/VatsalP117/iris/pkg/api"
	"github.com/VatsalP117/iris/pkg/db"
)

func main() {
	port := getEnv("PORT", "8080")
	dbPath := getEnv("DB_PATH", "./data/iris.db")

	if err := os.MkdirAll("data", 0755); err != nil {
		log.Fatal("Failed to create data directory:", err)
	}

	sqliteRepo, err := db.NewSqliteDB(dbPath)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer sqliteRepo.Close()
	if os.Getenv("IRIS_LAB_PPROF") == "1" {
		if rawExtraPages := os.Getenv("IRIS_LAB_DB_EXTRA_PAGES"); rawExtraPages != "" {
			extraPages, parseErr := strconv.Atoi(rawExtraPages)
			if parseErr != nil {
				log.Fatalf("Invalid IRIS_LAB_DB_EXTRA_PAGES: %v", parseErr)
			}
			appliedLimit, limitErr := sqliteRepo.ConstrainGrowthPages(context.Background(), extraPages)
			if limitErr != nil {
				log.Fatalf("Failed to constrain lab database: %v", limitErr)
			}
			log.Printf("Iris Reliability Lab database page limit: %d", appliedLimit)
		} else if resetErr := sqliteRepo.ResetGrowthPageLimit(context.Background()); resetErr != nil {
			log.Fatalf("Failed to reset lab database page limit: %v", resetErr)
		}
	}

	handler := api.NewHandler(sqliteRepo)
	mux := http.NewServeMux()

	mux.HandleFunc("/api/event", api.NewCORSMiddleware(handler.TrackEvent))
	mux.HandleFunc("/api/events", api.NewCORSMiddleware(handler.TrackBatchEvents))

	mux.HandleFunc("/api/stats", api.NewCORSMiddleware(handler.GetStats))
	mux.HandleFunc("/api/site-trends", api.NewCORSMiddleware(handler.GetSiteTrends))
	mux.HandleFunc("/api/pages", api.NewCORSMiddleware(handler.GetPages))
	mux.HandleFunc("/api/referrers", api.NewCORSMiddleware(handler.GetReferrers))
	mux.HandleFunc("/api/vitals", api.NewCORSMiddleware(handler.GetVitals))
	mux.HandleFunc("/api/vitals/distribution", api.NewCORSMiddleware(handler.GetVitalDistributions))
	mux.HandleFunc("/api/vitals/pages", api.NewCORSMiddleware(handler.GetPagePerformance))
	mux.HandleFunc("/api/vitals/score", api.NewCORSMiddleware(handler.GetPerformanceScore))
	mux.HandleFunc("/api/custom-events", api.NewCORSMiddleware(handler.GetCustomEvents))
	mux.HandleFunc("/api/custom-events/timeseries", api.NewCORSMiddleware(handler.GetCustomEventTimeSeries))
	mux.HandleFunc("/api/devices", api.NewCORSMiddleware(handler.GetDevices))
	mux.HandleFunc("/api/timeseries", api.NewCORSMiddleware(handler.GetTimeSeries))
	mux.HandleFunc("/api/timeseries/visitors", api.NewCORSMiddleware(handler.GetUniqueVisitorsTimeSeries))
	mux.HandleFunc("/api/timeseries/sessions", api.NewCORSMiddleware(handler.GetSessionsTimeSeries))
	mux.HandleFunc("/api/sites", api.NewCORSMiddleware(handler.ListSites))

	if os.Getenv("IRIS_LAB_PPROF") == "1" {
		mux.Handle("/debug/pprof/", http.DefaultServeMux)
		log.Printf("Iris Reliability Lab profiling enabled on /debug/pprof/")
	}

	dashboardDir := getEnv("DASHBOARD_DIR", "./dashboard/dist")
	fs := http.FileServer(http.Dir(dashboardDir))
	mux.Handle("/", fs)

	log.Printf("Iris Analytics listening on :%s (DB: %s)", port, dbPath)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
