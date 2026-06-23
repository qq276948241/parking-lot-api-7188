package main

import (
	"fmt"
	"log"
	"net/http"
)

const (
	DefaultPort       = 8080
	DefaultTotalSpace = 100
)

func main() {
	store := NewStore(DefaultTotalSpace)
	handler := NewHandler(store)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/entry", handler.HandleEntry)
	mux.HandleFunc("POST /api/exit", handler.HandleExit)
	mux.HandleFunc("GET /api/spaces", handler.HandleSpaces)
	mux.HandleFunc("GET /api/admin/income", handler.HandleDailyIncome)
	mux.HandleFunc("GET /api/admin/vehicles", handler.HandleActiveVehicles)
	mux.HandleFunc("POST /api/monthly", handler.HandleAddMonthly)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintln(w, "停车场管理系统 API 运行中")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "接口列表:")
		fmt.Fprintln(w, "  POST   /api/entry           车辆入场")
		fmt.Fprintln(w, "  POST   /api/exit            车辆出场")
		fmt.Fprintln(w, "  GET    /api/spaces           车位查询")
		fmt.Fprintln(w, "  GET    /api/admin/income     当日收入流水")
		fmt.Fprintln(w, "  GET    /api/admin/vehicles   在场车辆列表")
		fmt.Fprintln(w, "  POST   /api/monthly          办理月卡")
	})

	addr := fmt.Sprintf(":%d", DefaultPort)
	log.Printf("停车场管理系统启动，端口 %s，总车位 %d", addr, DefaultTotalSpace)
	log.Printf("计费规则: 首1小时%.0f元, 之后每小时%.0f元, 每日封顶%.0f元, %d分钟内免费",
		FirstHourFee, AdditionalHourFee, DailyMaxFee, FreeMinutes)
	log.Fatal(http.ListenAndServe(addr, mux))
}
