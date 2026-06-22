package service

import (
	"math"
	"parking-lot/model"
	"time"
)

const (
	FreeMinutes    = 30
	TempHourlyRate = 5.0
	TempDailyCap   = 50.0
)

type BillingService struct{}

func NewBillingService() *BillingService {
	return &BillingService{}
}

func (s *BillingService) CalculateFee(carType model.CarType, entryTime time.Time) (int64, float64) {
	now := time.Now()
	durationMin := int64(math.Ceil(now.Sub(entryTime).Minutes()))
	fee := s.calculateFee(carType, durationMin)
	return durationMin, fee
}

func (s *BillingService) calculateFee(carType model.CarType, durationMin int64) float64 {
	if carType == model.CarTypeMonthly {
		return 0
	}
	if durationMin <= FreeMinutes {
		return 0
	}
	billableMin := durationMin - FreeMinutes
	hours := math.Ceil(float64(billableMin) / 60.0)
	fee := hours * TempHourlyRate
	if fee > TempDailyCap {
		fee = TempDailyCap
	}
	return math.Round(fee*100) / 100
}
