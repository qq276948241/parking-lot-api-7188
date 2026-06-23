package main

import "time"

type VehicleType string

const (
	VehicleTypeTemporary VehicleType = "temporary"
	VehicleTypeMonthly   VehicleType = "monthly"
)

type ParkingRecord struct {
	ID          string      `json:"id"`
	PlateNumber string      `json:"plate_number"`
	VehicleType VehicleType `json:"vehicle_type"`
	EntryTime   time.Time   `json:"entry_time"`
	ExitTime    *time.Time  `json:"exit_time,omitempty"`
	Fee         float64     `json:"fee,omitempty"`
	IsParked    bool        `json:"is_parked"`
}

type ParkingLot struct {
	TotalSpaces int
	UsedSpaces  int
}

type DailyRevenue struct {
	Date    string          `json:"date"`
	Records []ParkingRecord `json:"records"`
	Total   float64         `json:"total"`
}
