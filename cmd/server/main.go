package main

import (
	"log"
	"net/http"
	"os"

	"github.com/VatsalP117/iris/pkg/api"
	"github.com/VatsalP117/iris/pkg/db"
)

func main() {
	port := getEnv("PORT", "8080")
	dbPath := getEnv("DB_PATH", "./data/iris.db")

	os.MkdirAll("data", 0755)

	sqliteRepo, err := db.NewSqliteDB(dbPath)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer sqliteRepo.Close()

	handler := api.NewHandler(sqliteRepo)

	// Event ingestion
	http.HandleFunc("/api/event", handler.TrackEvent)

	// Dashboard query API
	http.HandleFunc("/api/stats", handler.GetStats)
	http.HandleFunc("/api/pages", handler.GetPages)
	http.HandleFunc("/api/referrers", handler.GetReferrers)
	http.HandleFunc("/api/vitals", handler.GetVitals)
	http.HandleFunc("/api/devices", handler.GetDevices)
	http.HandleFunc("/api/timeseries", handler.GetTimeSeries)
	http.HandleFunc("/api/sites", handler.ListSites)

	// Serve dashboard static files
	dashboardDir := getEnv("DASHBOARD_DIR", "./dashboard/dist")
	fs := http.FileServer(http.Dir(dashboardDir))
	http.Handle("/", fs)

	log.Printf("Iris Analytics listening on :%s (DB: %s)", port, dbPath)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
