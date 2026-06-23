package main

import (
	"encoding/json"
	"net/http"
	"time"
)

type Handler struct {
	store *Store
}

func NewHandler(store *Store) *Handler {
	return &Handler{store: store}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (h *Handler) HandleEntry(w http.ResponseWriter, r *http.Request) {
	var req EntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "请求参数错误",
		})
		return
	}

	if req.Plate == "" {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "车牌号不能为空",
		})
		return
	}

	if req.VehicleType != VehicleTypeTemporary && req.VehicleType != VehicleTypeMonthly {
		req.VehicleType = VehicleTypeTemporary
	}

	record, err := h.store.Entry(req.Plate, req.VehicleType)
	if err != nil {
		writeJSON(w, http.StatusConflict, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Message: "入场成功",
		Data:    record,
	})
}

func (h *Handler) HandleExit(w http.ResponseWriter, r *http.Request) {
	var req ExitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "请求参数错误",
		})
		return
	}

	if req.Plate == "" {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "车牌号不能为空",
		})
		return
	}

	record, fee, durationMin, err := h.store.Exit(req.Plate)
	if err != nil {
		writeJSON(w, http.StatusNotFound, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "出场结算成功",
		Data: ExitResponse{
			Record:      record,
			Fee:         fee,
			DurationMin: durationMin,
		},
	})
}

func (h *Handler) HandleSpaces(w http.ResponseWriter, r *http.Request) {
	spaces := h.store.GetSpaces()
	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    spaces,
	})
}

func (h *Handler) HandleDailyIncome(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	var date time.Time
	if dateStr != "" {
		var err error
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "日期格式错误，请使用 YYYY-MM-DD",
			})
			return
		}
	} else {
		now := time.Now()
		date = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	}

	income := h.store.GetDailyIncome(date)
	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    income,
	})
}

func (h *Handler) HandleActiveVehicles(w http.ResponseWriter, r *http.Request) {
	vehicles := h.store.GetActiveVehicles()
	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    vehicles,
	})
}

func (h *Handler) HandleAddMonthly(w http.ResponseWriter, r *http.Request) {
	var req AddMonthlyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "请求参数错误",
		})
		return
	}

	if req.Plate == "" {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "车牌号不能为空",
		})
		return
	}

	if req.Duration <= 0 {
		req.Duration = 30
	}

	sub, err := h.store.AddMonthly(req.Plate, req.Duration)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Message: "月卡办理成功",
		Data:    sub,
	})
}
