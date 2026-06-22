package handler

import (
	"encoding/json"
	"net/http"
	"parking-lot/model"
	"parking-lot/service"
)

type CardHandler struct {
	svc *service.CardService
}

func NewCardHandler(svc *service.CardService) *CardHandler {
	return &CardHandler{svc: svc}
}

func (h *CardHandler) RegisterCard(w http.ResponseWriter, r *http.Request) {
	var req model.MonthlyCardRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, model.Response{Code: 400, Message: "请求参数错误"})
		return
	}
	if req.LicensePlate == "" {
		writeJSON(w, http.StatusBadRequest, model.Response{Code: 400, Message: "车牌号不能为空"})
		return
	}
	if req.OwnerName == "" {
		writeJSON(w, http.StatusBadRequest, model.Response{Code: 400, Message: "车主姓名不能为空"})
		return
	}
	if req.Months <= 0 {
		writeJSON(w, http.StatusBadRequest, model.Response{Code: 400, Message: "月数必须大于0"})
		return
	}

	card, err := h.svc.RegisterMonthlyCard(req.LicensePlate, req.OwnerName, req.Months)
	if err != nil {
		writeJSON(w, http.StatusConflict, model.Response{Code: 409, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "注册成功", Data: card})
}

func (h *CardHandler) RenewCard(w http.ResponseWriter, r *http.Request) {
	var req model.MonthlyCardRenewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, model.Response{Code: 400, Message: "请求参数错误"})
		return
	}
	if req.LicensePlate == "" {
		writeJSON(w, http.StatusBadRequest, model.Response{Code: 400, Message: "车牌号不能为空"})
		return
	}
	if req.Months <= 0 {
		writeJSON(w, http.StatusBadRequest, model.Response{Code: 400, Message: "续费月数必须大于0"})
		return
	}

	card, err := h.svc.RenewMonthlyCard(req.LicensePlate, req.Months)
	if err != nil {
		writeJSON(w, http.StatusNotFound, model.Response{Code: 404, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "续费成功", Data: card})
}

func (h *CardHandler) GetCard(w http.ResponseWriter, r *http.Request) {
	plate := r.URL.Query().Get("license_plate")
	if plate == "" {
		writeJSON(w, http.StatusBadRequest, model.Response{Code: 400, Message: "license_plate 参数不能为空"})
		return
	}

	card, err := h.svc.GetMonthlyCard(plate)
	if err != nil {
		writeJSON(w, http.StatusNotFound, model.Response{Code: 404, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: card})
}
