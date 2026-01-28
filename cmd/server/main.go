package main

import (
	"log"
	"net/http"
	"os"

	"github.com/VatsalP117/iris/pkg/api"
	"github.com/VatsalP117/iris/pkg/db"
)

func main() {
	os.Mkdir("data", 0755) 
	
	sqliteRepo, err := db.NewSqliteDB("./data/iris.db")
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer sqliteRepo.Close()

	handler := api.NewHandler(sqliteRepo)
	http.HandleFunc("/api/event", handler.TrackEvent)

	port := ":8080"
	log.Println("Pulse Analytics listening on", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}