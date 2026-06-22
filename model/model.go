package model

import "time"

type CarType string

const (
	CarTypeTemp     CarType = "temp"
	CarTypeMonthly  CarType = "monthly"
)

type VehicleRecord struct {
	LicensePlate string    `json:"license_plate"`
	CarType      CarType   `json:"car_type"`
	EntryTime    time.Time `json:"entry_time"`
}

type IncomeRecord struct {
	LicensePlate string    `json:"license_plate"`
	CarType      CarType   `json:"car_type"`
	EntryTime    time.Time `json:"entry_time"`
	ExitTime     time.Time `json:"exit_time"`
	DurationMin  int64     `json:"duration_min"`
	Fee          float64   `json:"fee"`
	CreatedAt    time.Time `json:"created_at"`
}

type EntryRequest struct {
	LicensePlate string `json:"license_plate"`
	CarType      CarType `json:"car_type"`
}

type ExitRequest struct {
	LicensePlate string `json:"license_plate"`
}

type ExitResponse struct {
	LicensePlate string    `json:"license_plate"`
	CarType      CarType   `json:"car_type"`
	EntryTime    time.Time `json:"entry_time"`
	ExitTime     time.Time `json:"exit_time"`
	DurationMin  int64     `json:"duration_min"`
	Fee          float64   `json:"fee"`
}

type SpacesResponse struct {
	Total      int `json:"total"`
	Occupied   int `json:"occupied"`
	Available  int `json:"available"`
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
