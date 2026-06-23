package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

type EntryRequest struct {
	PlateNumber string      `json:"plate_number"`
	VehicleType VehicleType `json:"vehicle_type"`
}

type ExitRequest struct {
	PlateNumber string `json:"plate_number"`
}

type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type Handler struct {
	store *Store
}

func NewHandler(store *Store) *Handler {
	return &Handler{store: store}
}

func writeJSON(w http.ResponseWriter, code int, resp APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Entry(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Code: 405, Message: "方法不允许"})
		return
	}

	var req EntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Code: 400, Message: "请求参数错误"})
		return
	}

	req.PlateNumber = strings.TrimSpace(req.PlateNumber)
	if req.PlateNumber == "" {
		writeJSON(w, http.StatusBadRequest, APIResponse{Code: 400, Message: ErrInvalidPlateNumber.Error()})
		return
	}

	if !validateVehicleType(req.VehicleType) {
		writeJSON(w, http.StatusBadRequest, APIResponse{Code: 400, Message: ErrInvalidVehicleType.Error()})
		return
	}

	record, err := h.store.Entry(req.PlateNumber, req.VehicleType)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Code: 400, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{Code: 0, Message: "入场成功", Data: record})
}

func (h *Handler) Exit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Code: 405, Message: "方法不允许"})
		return
	}

	var req ExitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Code: 400, Message: "请求参数错误"})
		return
	}

	req.PlateNumber = strings.TrimSpace(req.PlateNumber)
	if req.PlateNumber == "" {
		writeJSON(w, http.StatusBadRequest, APIResponse{Code: 400, Message: ErrInvalidPlateNumber.Error()})
		return
	}

	record, err := h.store.Exit(req.PlateNumber)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Code: 400, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{Code: 0, Message: "出场成功", Data: record})
}

func (h *Handler) ParkingStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Code: 405, Message: "方法不允许"})
		return
	}

	total, used := h.store.GetParkingLotStatus()
	data := map[string]interface{}{
		"total_spaces":   total,
		"used_spaces":    used,
		"available_spaces": total - used,
	}

	writeJSON(w, http.StatusOK, APIResponse{Code: 0, Message: "查询成功", Data: data})
}

func (h *Handler) ParkedVehicles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Code: 405, Message: "方法不允许"})
		return
	}

	vehicles := h.store.GetParkedVehicles()
	writeJSON(w, http.StatusOK, APIResponse{Code: 0, Message: "查询成功", Data: vehicles})
}

func (h *Handler) DailyRevenue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Code: 405, Message: "方法不允许"})
		return
	}

	date := r.URL.Query().Get("date")
	revenue := h.store.GetDailyRevenue(date)

	writeJSON(w, http.StatusOK, APIResponse{Code: 0, Message: "查询成功", Data: revenue})
}

func (h *Handler) VehicleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Code: 405, Message: "方法不允许"})
		return
	}

	plate := r.URL.Query().Get("plate_number")
	plate = strings.TrimSpace(plate)
	if plate == "" {
		writeJSON(w, http.StatusBadRequest, APIResponse{Code: 400, Message: ErrInvalidPlateNumber.Error()})
		return
	}

	record, exists := h.store.GetVehicleByPlate(plate)
	if !exists {
		writeJSON(w, http.StatusNotFound, APIResponse{Code: 404, Message: "车辆未在场内"})
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{Code: 0, Message: "查询成功", Data: record})
}

func SetupRoutes(handler *Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/entry", handler.Entry)
	mux.HandleFunc("/api/exit", handler.Exit)
	mux.HandleFunc("/api/parking/status", handler.ParkingStatus)
	mux.HandleFunc("/api/vehicle/query", handler.VehicleQuery)
	mux.HandleFunc("/api/admin/parked-vehicles", handler.ParkedVehicles)
	mux.HandleFunc("/api/admin/daily-revenue", handler.DailyRevenue)

	return mux
}
