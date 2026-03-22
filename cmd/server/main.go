package main

import (
	"log"
	"net/http"
	"os"
	"strings"

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

	ingestOrigins := getListEnv("IRIS_ALLOWED_INGEST_ORIGINS")
	dashboardOrigins := getListEnv("IRIS_ALLOWED_DASHBOARD_ORIGINS")

	log.Printf("Iris ingest CORS origins: %v", displayOriginConfig(ingestOrigins))
	log.Printf("Iris dashboard CORS origins: %v", displayOriginConfig(dashboardOrigins))

	ingestCORS := api.NewCORSMiddleware(api.CORSOptions{
		AllowedOrigins: ingestOrigins,
		AllowedMethods: []string{http.MethodPost, http.MethodOptions},
	})
	dashboardCORS := api.NewCORSMiddleware(api.CORSOptions{
		AllowedOrigins: dashboardOrigins,
		AllowedMethods: []string{http.MethodGet, http.MethodOptions},
	})

	http.HandleFunc("/api/event", ingestCORS(handler.TrackEvent))
	http.HandleFunc("/api/events", ingestCORS(handler.TrackBatchEvents))

	http.HandleFunc("/api/stats", dashboardCORS(handler.GetStats))
	http.HandleFunc("/api/pages", dashboardCORS(handler.GetPages))
	http.HandleFunc("/api/referrers", dashboardCORS(handler.GetReferrers))
	http.HandleFunc("/api/vitals", dashboardCORS(handler.GetVitals))
	http.HandleFunc("/api/devices", dashboardCORS(handler.GetDevices))
	http.HandleFunc("/api/timeseries", dashboardCORS(handler.GetTimeSeries))
	http.HandleFunc("/api/sites", dashboardCORS(handler.ListSites))

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

func getListEnv(key string) []string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value == "" {
			continue
		}
		values = append(values, value)
	}
	return values
}

func displayOriginConfig(values []string) []string {
	if len(values) == 0 {
		return []string{"*"}
	}
	return values
}
