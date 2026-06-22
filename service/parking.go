package service

import (
	"parking-lot/model"
	"sync"
	"time"
)

const TotalSpaces = 100

type ParkingService struct {
	mu        sync.RWMutex
	vehicles  map[string]*model.VehicleRecord
	incomes   []model.IncomeRecord
	cardSvc   *CardService
	billingSvc *BillingService
}

func NewParkingService(cardSvc *CardService, billingSvc *BillingService) *ParkingService {
	return &ParkingService{
		vehicles:   make(map[string]*model.VehicleRecord),
		incomes:    make([]model.IncomeRecord, 0),
		cardSvc:    cardSvc,
		billingSvc: billingSvc,
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
	if carType == "" || carType == model.CarTypeTemp {
		if s.cardSvc.IsCardValid(plate) {
			carType = model.CarTypeMonthly
		} else if carType == "" {
			carType = model.CarTypeTemp
		}
	}
	if carType != model.CarTypeTemp && carType != model.CarTypeMonthly {
		return nil, ErrInvalidCarType
	}
	if carType == model.CarTypeMonthly {
		card, err := s.cardSvc.GetMonthlyCard(plate)
		if err != nil {
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

	carType := record.CarType
	if carType == model.CarTypeTemp && s.cardSvc.IsCardValid(plate) {
		carType = model.CarTypeMonthly
	}

	durationMin, fee := s.billingSvc.CalculateFee(carType, record.EntryTime)
	now := time.Now()
	delete(s.vehicles, plate)

	income := model.IncomeRecord{
		LicensePlate: plate,
		CarType:      carType,
		EntryTime:    record.EntryTime,
		ExitTime:     now,
		DurationMin:  durationMin,
		Fee:          fee,
		CreatedAt:    now,
	}
	s.incomes = append(s.incomes, income)

	return &model.ExitResponse{
		LicensePlate: plate,
		CarType:      carType,
		EntryTime:    record.EntryTime,
		ExitTime:     now,
		DurationMin:  durationMin,
		Fee:          fee,
	}, nil
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
