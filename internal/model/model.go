package model

import "time"

type VehicleType string

const (
	VehicleTypeTemp    VehicleType = "temp"
	VehicleTypeMonthly VehicleType = "monthly"
)

type ParkingRecord struct {
	ID          string      `json:"id"`
	PlateNumber string      `json:"plate_number"`
	VehicleType VehicleType `json:"vehicle_type"`
	EntryTime   time.Time   `json:"entry_time"`
	ExitTime    *time.Time  `json:"exit_time,omitempty"`
	Fee         float64     `json:"fee,omitempty"`
	IsPaid      bool        `json:"is_paid"`
}

type ParkingLot struct {
	TotalSpots   int `json:"total_spots"`
	OccupiedSpots int `json:"occupied_spots"`
}

type DailyIncome struct {
	Date   string  `json:"date"`
	Total  float64 `json:"total"`
	Count  int     `json:"count"`
}

type PaymentRecord struct {
	ID          string    `json:"id"`
	PlateNumber string    `json:"plate_number"`
	Amount      float64   `json:"amount"`
	PayTime     time.Time `json:"pay_time"`
	VehicleType VehicleType `json:"vehicle_type"`
}
