package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"sync"
	"time"
)

type VehicleType string

const (
	TempVehicle    VehicleType = "temp"
	MonthlyVehicle VehicleType = "monthly"
)

type ParkingRecord struct {
	PlateNumber string      `json:"plateNumber"`
	VehicleType VehicleType `json:"vehicleType"`
	EntryTime   time.Time   `json:"entryTime"`
	ExitTime    *time.Time  `json:"exitTime,omitempty"`
	Fee         float64     `json:"fee"`
	IsPaid      bool        `json:"isPaid"`
}

type MonthlyCard struct {
	PlateNumber string    `json:"plateNumber"`
	ExpireTime  time.Time `json:"expireTime"`
}

type ParkingLot struct {
	mu            sync.RWMutex
	totalSpots    int
	records       map[string]*ParkingRecord
	history       []ParkingRecord
	monthlyCards  map[string]*MonthlyCard
}

func NewParkingLot(totalSpots int) *ParkingLot {
	return &ParkingLot{
		totalSpots:   totalSpots,
		records:      make(map[string]*ParkingRecord),
		history:      make([]ParkingRecord, 0),
		monthlyCards: make(map[string]*MonthlyCard),
	}
}

func (p *ParkingLot) AvailableSpots() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.totalSpots - len(p.records)
}

func (p *ParkingLot) Entry(plateNumber string, vehicleType VehicleType) (*ParkingRecord, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.records[plateNumber]; exists {
		return nil, fmt.Errorf("车辆已在场内")
	}

	if len(p.records) >= p.totalSpots {
		return nil, fmt.Errorf("车位已满")
	}

	if vehicleType == MonthlyVehicle {
		card, ok := p.monthlyCards[plateNumber]
		if !ok || card.ExpireTime.Before(time.Now()) {
			vehicleType = TempVehicle
		}
	}

	record := &ParkingRecord{
		PlateNumber: plateNumber,
		VehicleType: vehicleType,
		EntryTime:   time.Now(),
		IsPaid:      false,
	}

	p.records[plateNumber] = record
	return record, nil
}

func (p *ParkingLot) Exit(plateNumber string) (*ParkingRecord, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	record, exists := p.records[plateNumber]
	if !exists {
		return nil, fmt.Errorf("车辆不在场内")
	}

	exitTime := time.Now()
	record.ExitTime = &exitTime

	fee := p.calculateFee(record)
	record.Fee = fee
	record.IsPaid = true

	p.history = append(p.history, *record)
	delete(p.records, plateNumber)

	return record, nil
}

func (p *ParkingLot) calculateFee(record *ParkingRecord) float64 {
	if record.VehicleType == MonthlyVehicle {
		return 0
	}

	duration := record.ExitTime.Sub(record.EntryTime)
	hours := math.Ceil(duration.Hours())

	if hours <= 0 {
		return 0
	}

	firstHourRate := 5.0
	hourlyRate := 3.0
	dailyMax := 60.0

	days := int(hours) / 24
	remainingHours := int(hours) % 24

	var fee float64
	if days > 0 {
		fee += float64(days) * dailyMax
	}

	if remainingHours > 0 {
		if remainingHours == 1 {
			fee += firstHourRate
		} else {
			fee += firstHourRate + float64(remainingHours-1)*hourlyRate
		}
		if fee > float64(days+1)*dailyMax {
			fee = float64(days+1) * dailyMax
		}
	}

	return fee
}

func (p *ParkingLot) AddMonthlyCard(plateNumber string, durationDays int) (*MonthlyCard, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	expireTime := time.Now().AddDate(0, 0, durationDays)
	if existing, ok := p.monthlyCards[plateNumber]; ok {
		if existing.ExpireTime.After(time.Now()) {
			expireTime = existing.ExpireTime.AddDate(0, 0, durationDays)
		}
	}

	card := &MonthlyCard{
		PlateNumber: plateNumber,
		ExpireTime:  expireTime,
	}
	p.monthlyCards[plateNumber] = card
	return card, nil
}

func (p *ParkingLot) GetVehicles() []ParkingRecord {
	p.mu.RLock()
	defer p.mu.RUnlock()

	vehicles := make([]ParkingRecord, 0, len(p.records))
	for _, record := range p.records {
		vehicles = append(vehicles, *record)
	}
	return vehicles
}

func (p *ParkingLot) GetTodayIncome() (float64, []ParkingRecord) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	today := time.Now().Truncate(24 * time.Hour)
	var totalIncome float64
	records := make([]ParkingRecord, 0)

	for _, record := range p.history {
		if record.ExitTime != nil && record.ExitTime.After(today) {
			totalIncome += record.Fee
			records = append(records, record)
		}
	}

	return totalIncome, records
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func writeJSON(w http.ResponseWriter, code int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

type EntryRequest struct {
	PlateNumber string      `json:"plateNumber"`
	VehicleType VehicleType `json:"vehicleType"`
}

type ExitRequest struct {
	PlateNumber string `json:"plateNumber"`
}

type MonthlyCardRequest struct {
	PlateNumber string `json:"plateNumber"`
	Days        int    `json:"days"`
}

var lot *ParkingLot

func handleEntry(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, "方法不允许", nil)
		return
	}

	var req EntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, "请求参数错误", nil)
		return
	}

	if req.PlateNumber == "" {
		writeJSON(w, http.StatusBadRequest, "车牌号不能为空", nil)
		return
	}

	if req.VehicleType != TempVehicle && req.VehicleType != MonthlyVehicle {
		writeJSON(w, http.StatusBadRequest, "车辆类型错误", nil)
		return
	}

	record, err := lot.Entry(req.PlateNumber, req.VehicleType)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	writeJSON(w, http.StatusOK, "入场成功", record)
}

func handleExit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, "方法不允许", nil)
		return
	}

	var req ExitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, "请求参数错误", nil)
		return
	}

	if req.PlateNumber == "" {
		writeJSON(w, http.StatusBadRequest, "车牌号不能为空", nil)
		return
	}

	record, err := lot.Exit(req.PlateNumber)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	writeJSON(w, http.StatusOK, "出场成功", record)
}

func handleSpots(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, "方法不允许", nil)
		return
	}

	data := map[string]int{
		"total":     lot.totalSpots,
		"available": lot.AvailableSpots(),
		"occupied":  lot.totalSpots - lot.AvailableSpots(),
	}

	writeJSON(w, http.StatusOK, "查询成功", data)
}

func handleAdminVehicles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, "方法不允许", nil)
		return
	}

	vehicles := lot.GetVehicles()
	writeJSON(w, http.StatusOK, "查询成功", vehicles)
}

func handleAdminIncome(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, "方法不允许", nil)
		return
	}

	totalIncome, records := lot.GetTodayIncome()
	data := map[string]interface{}{
		"totalIncome": totalIncome,
		"records":     records,
	}

	writeJSON(w, http.StatusOK, "查询成功", data)
}

func handleAddMonthlyCard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, "方法不允许", nil)
		return
	}

	var req MonthlyCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, "请求参数错误", nil)
		return
	}

	if req.PlateNumber == "" {
		writeJSON(w, http.StatusBadRequest, "车牌号不能为空", nil)
		return
	}

	if req.Days <= 0 {
		writeJSON(w, http.StatusBadRequest, "天数必须大于0", nil)
		return
	}

	card, err := lot.AddMonthlyCard(req.PlateNumber, req.Days)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	writeJSON(w, http.StatusOK, "月卡添加成功", card)
}

func main() {
	lot = NewParkingLot(100)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/entry", handleEntry)
	mux.HandleFunc("/api/exit", handleExit)
	mux.HandleFunc("/api/spots", handleSpots)
	mux.HandleFunc("/api/admin/income", handleAdminIncome)
	mux.HandleFunc("/api/admin/vehicles", handleAdminVehicles)
	mux.HandleFunc("/api/admin/monthly", handleAddMonthlyCard)

	addr := ":8080"
	log.Printf("停车场管理系统启动，监听 %s", addr)
	log.Printf("总车位: %d", lot.totalSpots)
	log.Fatal(http.ListenAndServe(addr, mux))
}
