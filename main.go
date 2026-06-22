package main

import (
	"fmt"
	"log"
	"net/http"
	"parking-api/internal/handler"
	"parking-api/internal/service"
	"parking-api/internal/store"
)

func main() {
	totalSpots := 100
	s := store.NewMemoryStore(totalSpots)

	s.AddMonthlyPlate("京A12345")
	s.AddMonthlyPlate("京B88888")

	billing := service.NewBillingService(service.DefaultPricingConfig())
	h := handler.NewHandler(s, billing)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/entry", handler.WithMethod(http.MethodPost, h.Entry))
	mux.HandleFunc("/api/exit", handler.WithMethod(http.MethodPost, h.Exit))
	mux.HandleFunc("/api/parking/status", handler.WithMethod(http.MethodGet, h.ParkingLotStatus))
	mux.HandleFunc("/api/vehicle/query", handler.WithMethod(http.MethodGet, h.QueryVehicle))
	mux.HandleFunc("/api/admin/active-vehicles", handler.WithMethod(http.MethodGet, h.ActiveVehicles))
	mux.HandleFunc("/api/admin/today-income", handler.WithMethod(http.MethodGet, h.TodayIncome))

	addr := ":8080"
	fmt.Printf("Parking API server starting on %s\n", addr)
	fmt.Println("Endpoints:")
	fmt.Println("  POST   /api/entry                    - 车辆入场")
	fmt.Println("  POST   /api/exit                     - 车辆出场")
	fmt.Println("  GET    /api/parking/status           - 车位状态查询")
	fmt.Println("  GET    /api/vehicle/query               - 按车牌号查询车辆状态")
	fmt.Println("  GET    /api/admin/active-vehicles    - 在场车辆列表")
	fmt.Println("  GET    /api/admin/today-income       - 当日收入流水")

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
