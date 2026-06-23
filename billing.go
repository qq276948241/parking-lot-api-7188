package main

import (
	"crypto/rand"
	"encoding/hex"
	"math"
	"time"
)

const (
	TemporaryRatePerHour = 5.0
	TemporaryMaxFee      = 50.0
	MonthlyFreeHours     = 24
	MonthlyExtraRate     = 2.0
	MonthlyExtraMaxFee   = 20.0
)

func calculateFee(vehicleType VehicleType, entryTime, exitTime time.Time) float64 {
	duration := exitTime.Sub(entryTime)
	hours := duration.Hours()

	if hours <= 0.5 {
		return 0.0
	}

	if vehicleType == VehicleTypeTemporary {
		fee := math.Ceil(hours) * TemporaryRatePerHour
		if fee > TemporaryMaxFee {
			return TemporaryMaxFee
		}
		return fee
	}

	if hours <= MonthlyFreeHours {
		return 0.0
	}

	extraHours := hours - MonthlyFreeHours
	fee := math.Ceil(extraHours) * MonthlyExtraRate
	if fee > MonthlyExtraMaxFee {
		return MonthlyExtraMaxFee
	}
	return fee
}

func generateID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return time.Now().Format("20060102150405")
	}
	return hex.EncodeToString(b)
}

func validateVehicleType(vt VehicleType) bool {
	return vt == VehicleTypeTemporary || vt == VehicleTypeMonthly
}
