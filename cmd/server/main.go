package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

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
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "rebuild-projections":
			if err := sqliteRepo.RebuildProjections(context.Background()); err != nil {
				log.Fatalf("Failed to rebuild projections: %v", err)
			}
			log.Print("Iris analytics projections rebuilt")
			return
		case "apply-retention":
			deleted, err := sqliteRepo.ApplyRetention(context.Background(), time.Now().UTC())
			if err != nil {
				log.Fatalf("Failed to apply retention: %v", err)
			}
			log.Printf("Iris retention removed %d expired raw events", deleted)
			return
		default:
			log.Fatalf("Unknown command %q", os.Args[1])
		}
	}
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
	mux.HandleFunc("/api/sites", api.NewCORSMiddleware(handler.Sites))
	mux.HandleFunc("/api/status", api.NewCORSMiddleware(handler.Status))
	mux.HandleFunc("/healthz", handler.Status)

	if os.Getenv("IRIS_LAB_PPROF") == "1" {
		mux.Handle("/debug/pprof/", http.DefaultServeMux)
		log.Printf("Iris Reliability Lab profiling enabled on /debug/pprof/")
	}

	dashboardDir := getEnv("DASHBOARD_DIR", "./dashboard/dist")
	fs := http.FileServer(http.Dir(dashboardDir))
	mux.Handle("/", fs)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go runMaintenance(ctx, sqliteRepo)

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       2 * time.Minute,
	}
	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("Iris Analytics listening on :%s (DB: %s)", port, dbPath)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Iris shutdown error: %v", err)
		}
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Iris server error: %v", err)
		}
	}
}

func runMaintenance(ctx context.Context, repo *db.SqliteRepository) {
	projectorTicker := time.NewTicker(250 * time.Millisecond)
	retentionTicker := time.NewTicker(24 * time.Hour)
	defer projectorTicker.Stop()
	defer retentionTicker.Stop()

	project := func() {
		for {
			count, err := repo.ProjectPending(ctx, 1000)
			if err != nil {
				if ctx.Err() == nil {
					log.Printf("[Projector] error: %v", err)
				}
				return
			}
			if count < 1000 {
				return
			}
		}
	}
	applyRetention := func() {
		deleted, err := repo.ApplyRetention(ctx, time.Now().UTC())
		if err != nil {
			if ctx.Err() == nil {
				log.Printf("[Retention] error: %v", err)
			}
			return
		}
		if deleted > 0 {
			log.Printf("[Retention] removed %d expired raw events", deleted)
		}
	}

	project()
	applyRetention()
	for {
		select {
		case <-ctx.Done():
			return
		case <-projectorTicker.C:
			project()
		case <-retentionTicker.C:
			applyRetention()
		}
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
