package main

import "time"

type VehicleType string

const (
	VehicleTypeTemporary VehicleType = "temporary"
	VehicleTypeMonthly   VehicleType = "monthly"
)

type ParkingRecord struct {
	ID          string      `json:"id"`
	Plate       string      `json:"plate"`
	VehicleType VehicleType `json:"vehicle_type"`
	EntryTime   time.Time   `json:"entry_time"`
	ExitTime    *time.Time  `json:"exit_time,omitempty"`
	Fee         float64     `json:"fee"`
	Paid        bool        `json:"paid"`
}

type MonthlySubscription struct {
	Plate     string    `json:"plate"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

type EntryRequest struct {
	Plate       string      `json:"plate"`
	VehicleType VehicleType `json:"vehicle_type"`
}

type ExitRequest struct {
	Plate string `json:"plate"`
}

type ExitResponse struct {
	Record      *ParkingRecord `json:"record"`
	Fee         float64        `json:"fee"`
	DurationMin int64          `json:"duration_min"`
}

type SpacesResponse struct {
	Total     int `json:"total"`
	Occupied  int `json:"occupied"`
	Available int `json:"available"`
}

type IncomeRecord struct {
	Plate       string      `json:"plate"`
	VehicleType VehicleType `json:"vehicle_type"`
	EntryTime   time.Time   `json:"entry_time"`
	ExitTime    time.Time   `json:"exit_time"`
	Fee         float64     `json:"fee"`
}

type IncomeResponse struct {
	Date      string         `json:"date"`
	TotalFee  float64        `json:"total_fee"`
	Count     int            `json:"count"`
	Records   []IncomeRecord `json:"records"`
}

type AddMonthlyRequest struct {
	Plate    string `json:"plate"`
	Duration int    `json:"duration_days"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}
