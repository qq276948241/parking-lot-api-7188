package service

import (
	"math"
	"parking-lot/model"
	"sync"
	"time"
)

const (
	TotalSpaces     = 100
	FreeMinutes     = 30
	TempHourlyRate  = 5.0
	TempDailyCap    = 50.0
)

type ParkingService struct {
	mu       sync.RWMutex
	vehicles map[string]*model.VehicleRecord
	incomes  []model.IncomeRecord
	cards    map[string]*model.MonthlyCard
}

func NewParkingService() *ParkingService {
	return &ParkingService{
		vehicles: make(map[string]*model.VehicleRecord),
		incomes:  make([]model.IncomeRecord, 0),
		cards:    make(map[string]*model.MonthlyCard),
	}
}

func (s *ParkingService) VehicleEntry(plate string, carType model.CarType) (*model.VehicleRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.vehicles[plate]; exists {
		return nil, ErrVehicleAlreadyIn
	}
	if len(s.vehicles) >= TotalSpaces {
		return nil, ErrNoSpace
	}
	if carType != model.CarTypeTemp && carType != model.CarTypeMonthly {
		return nil, ErrInvalidCarType
	}

	if carType == model.CarTypeMonthly {
		card, exists := s.cards[plate]
		if !exists {
			return nil, ErrMonthlyCardNotFound
		}
		if !card.Active || time.Now().After(card.ExpireDate) {
			return nil, ErrMonthlyCardExpired
		}
	}

	record := &model.VehicleRecord{
		LicensePlate: plate,
		CarType:      carType,
		EntryTime:    time.Now(),
	}
	s.vehicles[plate] = record
	return record, nil
}

func (s *ParkingService) VehicleExit(plate string) (*model.ExitResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, exists := s.vehicles[plate]
	if !exists {
		return nil, ErrVehicleNotFound
	}

	now := time.Now()
	durationMin := int64(math.Ceil(now.Sub(record.EntryTime).Minutes()))
	fee := s.calculateFee(record.CarType, durationMin)

	delete(s.vehicles, plate)

	income := model.IncomeRecord{
		LicensePlate: plate,
		CarType:      record.CarType,
		EntryTime:    record.EntryTime,
		ExitTime:     now,
		DurationMin:  durationMin,
		Fee:          fee,
		CreatedAt:    now,
	}
	s.incomes = append(s.incomes, income)

	return &model.ExitResponse{
		LicensePlate: plate,
		CarType:      record.CarType,
		EntryTime:    record.EntryTime,
		ExitTime:     now,
		DurationMin:  durationMin,
		Fee:          fee,
	}, nil
}

func (s *ParkingService) calculateFee(carType model.CarType, durationMin int64) float64 {
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

func (s *ParkingService) GetSpaces() model.SpacesResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	occupied := len(s.vehicles)
	return model.SpacesResponse{
		Total:     TotalSpaces,
		Occupied:  occupied,
		Available: TotalSpaces - occupied,
	}
}

func (s *ParkingService) GetTodayIncomes() []model.IncomeRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	today := time.Now().Truncate(24 * time.Hour)
	var result []model.IncomeRecord
	for _, r := range s.incomes {
		if r.CreatedAt.Truncate(24*time.Hour).Equal(today) {
			result = append(result, r)
		}
	}
	return result
}

func (s *ParkingService) GetParkedVehicles() []model.VehicleRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []model.VehicleRecord
	for _, v := range s.vehicles {
		result = append(result, *v)
	}
	return result
}

func (s *ParkingService) RegisterMonthlyCard(plate, ownerName string, months int) (*model.MonthlyCard, error) {
	if months <= 0 {
		return nil, ErrInvalidMonths
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.cards[plate]; exists {
		return nil, ErrMonthlyCardExists
	}

	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	expireDate := startDate.AddDate(0, months, 1)

	card := &model.MonthlyCard{
		LicensePlate: plate,
		OwnerName:    ownerName,
		StartDate:    startDate,
		ExpireDate:   expireDate,
		Active:       true,
		CreatedAt:    now,
	}
	s.cards[plate] = card
	return card, nil
}

func (s *ParkingService) RenewMonthlyCard(plate string, months int) (*model.MonthlyCard, error) {
	if months <= 0 {
		return nil, ErrInvalidMonths
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	card, exists := s.cards[plate]
	if !exists {
		return nil, ErrMonthlyCardNotFound
	}

	base := card.ExpireDate
	now := time.Now()
	if now.After(base) {
		base = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	}
	card.ExpireDate = base.AddDate(0, months, 0)
	card.Active = true
	return card, nil
}

func (s *ParkingService) GetMonthlyCard(plate string) (*model.MonthlyCard, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	card, exists := s.cards[plate]
	if !exists {
		return nil, ErrMonthlyCardNotFound
	}

	if card.Active && time.Now().After(card.ExpireDate) {
		cardCopy := *card
		cardCopy.Active = false
		return &cardCopy, nil
	}
	return card, nil
}
