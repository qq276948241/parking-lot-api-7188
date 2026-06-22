package handler

import (
	"encoding/json"
	"net/http"
	"parking-lot/model"
	"parking-lot/service"
)

type ParkingHandler struct {
	svc *service.ParkingService
}

func NewParkingHandler(svc *service.ParkingService) *ParkingHandler {
	return &ParkingHandler{svc: svc}
}

func writeJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func (h *ParkingHandler) Entry(w http.ResponseWriter, r *http.Request) {
	var req model.EntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, model.Response{Code: 400, Message: "请求参数错误"})
		return
	}
	if req.LicensePlate == "" {
		writeJSON(w, http.StatusBadRequest, model.Response{Code: 400, Message: "车牌号不能为空"})
		return
	}
	if req.CarType == "" {
		req.CarType = model.CarTypeTemp
	}

	record, err := h.svc.VehicleEntry(req.LicensePlate, req.CarType)
	if err != nil {
		writeJSON(w, http.StatusConflict, model.Response{Code: 409, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "入场成功", Data: record})
}

func (h *ParkingHandler) Exit(w http.ResponseWriter, r *http.Request) {
	var req model.ExitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, model.Response{Code: 400, Message: "请求参数错误"})
		return
	}
	if req.LicensePlate == "" {
		writeJSON(w, http.StatusBadRequest, model.Response{Code: 400, Message: "车牌号不能为空"})
		return
	}

	resp, err := h.svc.VehicleExit(req.LicensePlate)
	if err != nil {
		writeJSON(w, http.StatusNotFound, model.Response{Code: 404, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "出场成功", Data: resp})
}

func (h *ParkingHandler) Spaces(w http.ResponseWriter, r *http.Request) {
	spaces := h.svc.GetSpaces()
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: spaces})
}

func (h *ParkingHandler) TodayIncomes(w http.ResponseWriter, r *http.Request) {
	incomes := h.svc.GetTodayIncomes()
	if incomes == nil {
		incomes = []model.IncomeRecord{}
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: incomes})
}

func (h *ParkingHandler) ParkedVehicles(w http.ResponseWriter, r *http.Request) {
	vehicles := h.svc.GetParkedVehicles()
	if vehicles == nil {
		vehicles = []model.VehicleRecord{}
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: vehicles})
}
