package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"parking-api/internal/model"
	"parking-api/internal/service"
	"parking-api/internal/store"
	"time"

	"github.com/google/uuid"
)

type Handler struct {
	store   *store.MemoryStore
	billing *service.BillingService
}

func NewHandler(s *store.MemoryStore, b *service.BillingService) *Handler {
	return &Handler{store: s, billing: b}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func WithMethod(method string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		next(w, r)
	}
}

func decodeJSON(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

type EntryRequest struct {
	PlateNumber string           `json:"plate_number"`
	VehicleType model.VehicleType `json:"vehicle_type"`
}

func (h *Handler) Entry(w http.ResponseWriter, r *http.Request) {
	var req EntryRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.PlateNumber == "" {
		writeError(w, http.StatusBadRequest, "plate_number is required")
		return
	}

	vType := req.VehicleType
	if vType == "" {
		if h.store.IsMonthlyPlate(req.PlateNumber) {
			vType = model.VehicleTypeMonthly
		} else {
			vType = model.VehicleTypeTemp
		}
	}

	record := &model.ParkingRecord{
		ID:          uuid.New().String(),
		PlateNumber: req.PlateNumber,
		VehicleType: vType,
		EntryTime:   time.Now(),
		IsPaid:      false,
	}

	if err := h.store.Entry(record); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, record)
}

type ExitRequest struct {
	PlateNumber string `json:"plate_number"`
}

func (h *Handler) Exit(w http.ResponseWriter, r *http.Request) {
	var req ExitRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.PlateNumber == "" {
		writeError(w, http.StatusBadRequest, "plate_number is required")
		return
	}

	record, err := h.store.GetActiveRecordByPlate(req.PlateNumber)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	exitTime := time.Now()
	fee := h.billing.CalculateFee(record, exitTime)

	updated, err := h.store.Exit(record.ID, fee, exitTime)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":           updated.ID,
		"plate_number": updated.PlateNumber,
		"vehicle_type": updated.VehicleType,
		"entry_time":   updated.EntryTime,
		"exit_time":    updated.ExitTime,
		"duration_min": exitTime.Sub(updated.EntryTime).Minutes(),
		"fee":          updated.Fee,
	})
}

func (h *Handler) ParkingLotStatus(w http.ResponseWriter, r *http.Request) {
	status := h.store.GetParkingLot()
	remaining := status.TotalSpots - status.OccupiedSpots
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"total_spots":     status.TotalSpots,
		"occupied_spots":  status.OccupiedSpots,
		"available_spots": remaining,
	})
}

func (h *Handler) ActiveVehicles(w http.ResponseWriter, r *http.Request) {
	vehicles := h.store.GetActiveVehicles()
	if vehicles == nil {
		vehicles = []model.ParkingRecord{}
	}
	writeJSON(w, http.StatusOK, vehicles)
}

func (h *Handler) TodayIncome(w http.ResponseWriter, r *http.Request) {
	payments := h.store.GetTodayPayments()
	var total float64
	for _, p := range payments {
		total += p.Amount
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"date":     time.Now().Format("2006-01-02"),
		"total":    total,
		"count":    len(payments),
		"payments": payments,
	})
}

func (h *Handler) QueryVehicle(w http.ResponseWriter, r *http.Request) {
	plate := r.URL.Query().Get("plate_number")
	if plate == "" {
		writeError(w, http.StatusBadRequest, "plate_number is required")
		return
	}

	record, err := h.store.GetActiveRecordByPlate(plate)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"plate_number":  plate,
			"is_in_parking": false,
		})
		return
	}

	duration := time.Since(record.EntryTime)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"plate_number":   record.PlateNumber,
		"is_in_parking":  true,
		"vehicle_type":   record.VehicleType,
		"entry_time":     record.EntryTime,
		"duration_min":   duration.Minutes(),
		"duration_human": fmtDuration(duration),
	})
}

func fmtDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	if h > 0 {
		return fmt.Sprintf("%d小时%d分%d秒", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%d分%d秒", m, s)
	}
	return fmt.Sprintf("%d秒", s)
}
