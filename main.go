package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/zhangkui/urban-traffic-optimization/internal/api"
	"github.com/zhangkui/urban-traffic-optimization/internal/realtime"
	"github.com/zhangkui/urban-traffic-optimization/internal/store"
)

func main() {
	path := os.Getenv("URBAN_TRAFFIC_OPTIMIZATION_DB")
	if path == "" {
		path = "data/urban-traffic-optimization.db"
	}
	repository, err := store.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer repository.Close()
	hub := realtime.NewHub()
	go hub.Run(context.Background())
	address := os.Getenv("URBAN_TRAFFIC_OPTIMIZATION_ADDR")
	if address == "" {
		address = ":8080"
	}
	server := api.New(repository, hub)
	log.Printf("urban-traffic-optimization listening on %s", address)
	log.Fatal(http.ListenAndServe(address, server.Handler()))
}
