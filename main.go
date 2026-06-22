package main

import (
	"fmt"
	"log"
	"net/http"
	"parking-lot/handler"
	"parking-lot/service"
)

func main() {
	svc := service.NewParkingService()
	h := handler.NewParkingHandler(svc)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/entry", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		h.Entry(w, r)
	})

	mux.HandleFunc("/api/exit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		h.Exit(w, r)
	})

	mux.HandleFunc("/api/spaces", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		h.Spaces(w, r)
	})

	mux.HandleFunc("/api/admin/income", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		h.TodayIncomes(w, r)
	})

	mux.HandleFunc("/api/admin/vehicles", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		h.ParkedVehicles(w, r)
	})

	addr := ":8080"
	fmt.Printf("停车场管理 API 服务启动，监听 %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
