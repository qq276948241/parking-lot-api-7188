package service

import (
	"parking-api/internal/model"
	"time"
)

type PricingConfig struct {
	TempHourlyRate      float64
	TempDailyMax        float64
	TempFreeMinutes     int
}

func DefaultPricingConfig() PricingConfig {
	return PricingConfig{
		TempHourlyRate:  5.0,
		TempDailyMax:    60.0,
		TempFreeMinutes: 15,
	}
}

type BillingService struct {
	config PricingConfig
}

func NewBillingService(config PricingConfig) *BillingService {
	return &BillingService{config: config}
}

func (s *BillingService) CalculateFee(record *model.ParkingRecord, exitTime time.Time) float64 {
	if record.VehicleType == model.VehicleTypeMonthly {
		return 0.0
	}

	duration := exitTime.Sub(record.EntryTime)
	freeDuration := time.Duration(s.config.TempFreeMinutes) * time.Minute

	if duration <= freeDuration {
		return 0.0
	}

	billableDuration := duration - freeDuration
	hours := billableDuration.Hours()
	if billableDuration.Minutes() > 0 && billableDuration.Minutes() < 60 {
		hours = 1.0
	}

	totalDays := int(duration.Hours() / 24)
	remainingHours := duration.Hours() - float64(totalDays*24)

	var fee float64
	if totalDays > 0 {
		fee += float64(totalDays) * s.config.TempDailyMax
	}

	if remainingHours > 0 {
		remainingFee := remainingHours * s.config.TempHourlyRate
		if remainingFee > s.config.TempDailyMax {
			remainingFee = s.config.TempDailyMax
		}
		fee += remainingFee
	}

	if fee == 0 {
		fee = hours * s.config.TempHourlyRate
		if fee > s.config.TempDailyMax {
			fee = s.config.TempDailyMax
		}
	}

	return fee
}
