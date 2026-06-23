package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

const (
	DefaultTotalSpaces = 100
	DefaultPort        = 8080
)

func main() {
	totalSpaces := DefaultTotalSpaces
	if s := os.Getenv("TOTAL_SPACES"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 {
			totalSpaces = v
		}
	}

	port := DefaultPort
	if s := os.Getenv("PORT"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 {
			port = v
		}
	}

	store := NewStore(totalSpaces)
	handler := NewHandler(store)
	router := SetupRoutes(handler)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("停车场管理服务启动，端口: %d，总车位: %d", port, totalSpaces)
	log.Printf("API 接口:")
	log.Printf("  POST /api/entry              - 车辆入场")
	log.Printf("  POST /api/exit               - 车辆出场")
	log.Printf("  GET  /api/parking/status     - 车位状态查询")
	log.Printf("  GET  /api/vehicle/query      - 车辆查询")
	log.Printf("  GET  /api/admin/parked-vehicles - 在场车辆列表")
	log.Printf("  GET  /api/admin/daily-revenue   - 当日收入流水")

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
