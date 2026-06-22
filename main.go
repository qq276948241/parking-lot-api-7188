package main

import (
	"fmt"
	"log"
	"net/http"
	"parking-lot/handler"
	"parking-lot/service"
)

func main() {
	cardSvc := service.NewCardService()
	billingSvc := service.NewBillingService()
	parkingSvc := service.NewParkingService(cardSvc, billingSvc)

	parkingH := handler.NewParkingHandler(parkingSvc)
	cardH := handler.NewCardHandler(cardSvc)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/entry", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		parkingH.Entry(w, r)
	})

	mux.HandleFunc("/api/exit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		parkingH.Exit(w, r)
	})

	mux.HandleFunc("/api/spaces", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		parkingH.Spaces(w, r)
	})

	mux.HandleFunc("/api/admin/income", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		parkingH.TodayIncomes(w, r)
	})

	mux.HandleFunc("/api/admin/vehicles", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		parkingH.ParkedVehicles(w, r)
	})

	mux.HandleFunc("/api/card/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		cardH.RegisterCard(w, r)
	})

	mux.HandleFunc("/api/card/renew", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		cardH.RenewCard(w, r)
	})

	mux.HandleFunc("/api/card/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		cardH.GetCard(w, r)
	})

	addr := ":8080"
	fmt.Printf("停车场管理 API 服务启动，监听 %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
