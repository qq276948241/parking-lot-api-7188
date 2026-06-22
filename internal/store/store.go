package store

import (
	"errors"
	"parking-api/internal/model"
	"sync"
	"time"
)

type Store interface {
	AddMonthlyPlate(plate string)
	RemoveMonthlyPlate(plate string) bool
	IsMonthlyPlate(plate string) bool
	GetMonthlyPlates() []string
	GetParkingLot() *model.ParkingLot
	Entry(record *model.ParkingRecord) error
	GetActiveRecordByPlate(plate string) (*model.ParkingRecord, error)
	Exit(recordID string, fee float64, exitTime time.Time) (*model.ParkingRecord, error)
	GetActiveVehicles() []model.ParkingRecord
	GetTodayPayments() []model.PaymentRecord
}

type MemoryStore struct {
	mu              sync.RWMutex
	records         map[string]*model.ParkingRecord
	activePlates    map[string]bool
	payments        []model.PaymentRecord
	totalSpots      int
	monthlyPlates   map[string]bool
}

func NewMemoryStore(totalSpots int) *MemoryStore {
	return &MemoryStore{
		records:       make(map[string]*model.ParkingRecord),
		activePlates:  make(map[string]bool),
		payments:      []model.PaymentRecord{},
		totalSpots:    totalSpots,
		monthlyPlates: make(map[string]bool),
	}
}

func (s *MemoryStore) AddMonthlyPlate(plate string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.monthlyPlates[plate] = true
}

func (s *MemoryStore) IsMonthlyPlate(plate string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.monthlyPlates[plate]
}

func (s *MemoryStore) RemoveMonthlyPlate(plate string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.monthlyPlates[plate] {
		return false
	}
	delete(s.monthlyPlates, plate)
	return true
}

func (s *MemoryStore) GetMonthlyPlates() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	plates := make([]string, 0, len(s.monthlyPlates))
	for p := range s.monthlyPlates {
		plates = append(plates, p)
	}
	return plates
}

func (s *MemoryStore) GetParkingLot() *model.ParkingLot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return &model.ParkingLot{
		TotalSpots:    s.totalSpots,
		OccupiedSpots: len(s.activePlates),
	}
}

func (s *MemoryStore) Entry(record *model.ParkingRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.activePlates[record.PlateNumber] {
		return errors.New("vehicle already in parking lot")
	}

	if len(s.activePlates) >= s.totalSpots {
		return errors.New("parking lot is full")
	}

	s.records[record.ID] = record
	s.activePlates[record.PlateNumber] = true
	return nil
}

func (s *MemoryStore) GetActiveRecordByPlate(plate string) (*model.ParkingRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.activePlates[plate] {
		return nil, errors.New("vehicle not found in parking lot")
	}

	for _, r := range s.records {
		if r.PlateNumber == plate && r.ExitTime == nil {
			return r, nil
		}
	}
	return nil, errors.New("vehicle not found in parking lot")
}

func (s *MemoryStore) Exit(recordID string, fee float64, exitTime time.Time) (*model.ParkingRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.records[recordID]
	if !ok {
		return nil, errors.New("record not found")
	}
	if record.ExitTime != nil {
		return nil, errors.New("vehicle already exited")
	}

	record.ExitTime = &exitTime
	record.Fee = fee
	record.IsPaid = true
	delete(s.activePlates, record.PlateNumber)

	s.payments = append(s.payments, model.PaymentRecord{
		ID:          "PAY-" + recordID,
		PlateNumber: record.PlateNumber,
		Amount:      fee,
		PayTime:     exitTime,
		VehicleType: record.VehicleType,
	})

	return record, nil
}

func (s *MemoryStore) GetActiveVehicles() []model.ParkingRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	active := []model.ParkingRecord{}
	for _, r := range s.records {
		if r.ExitTime == nil {
			active = append(active, *r)
		}
	}
	return active
}

func (s *MemoryStore) GetTodayPayments() []model.PaymentRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	today := time.Now().Format("2006-01-02")
	result := []model.PaymentRecord{}
	for _, p := range s.payments {
		if p.PayTime.Format("2006-01-02") == today {
			result = append(result, p)
		}
	}
	return result
}
