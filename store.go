package main

import (
	"fmt"
	"sync"
	"time"
)

type Store struct {
	mu              sync.RWMutex
	records         map[string]*ParkingRecord
	activeRecords   map[string]*ParkingRecord
	monthlySubs     map[string]*MonthlySubscription
	totalSpaces     int
	idCounter       int
}

func NewStore(totalSpaces int) *Store {
	return &Store{
		records:       make(map[string]*ParkingRecord),
		activeRecords: make(map[string]*ParkingRecord),
		monthlySubs:   make(map[string]*MonthlySubscription),
		totalSpaces:   totalSpaces,
	}
}

func (s *Store) nextID() string {
	s.idCounter++
	return fmt.Sprintf("P%06d", s.idCounter)
}

func (s *Store) Entry(plate string, vType VehicleType) (*ParkingRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.activeRecords[plate]; exists {
		return nil, fmt.Errorf("车辆 %s 已在场内", plate)
	}

	if len(s.activeRecords) >= s.totalSpaces {
		return nil, fmt.Errorf("车位已满，无法入场")
	}

	if vType == VehicleTypeMonthly {
		if !s.isMonthlyValidLocked(plate) {
			return nil, fmt.Errorf("车辆 %s 月卡已过期或未办理，请以临时车入场", plate)
		}
	}

	record := &ParkingRecord{
		ID:          s.nextID(),
		Plate:       plate,
		VehicleType: vType,
		EntryTime:   time.Now(),
	}

	s.records[record.ID] = record
	s.activeRecords[plate] = record

	return record, nil
}

func (s *Store) Exit(plate string) (*ParkingRecord, float64, int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, exists := s.activeRecords[plate]
	if !exists {
		return nil, 0, 0, fmt.Errorf("车辆 %s 不在场内", plate)
	}

	now := time.Now()
	record.ExitTime = &now
	duration := now.Sub(record.EntryTime)
	durationMin := int64(duration.Minutes())

	var fee float64
	if record.VehicleType == VehicleTypeMonthly && s.isMonthlyValidLocked(plate) {
		fee = 0
	} else {
		fee = CalculateFee(duration)
	}

	record.Fee = fee
	record.Paid = true
	delete(s.activeRecords, plate)

	return record, fee, durationMin, nil
}

func (s *Store) GetSpaces() SpacesResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	occupied := len(s.activeRecords)
	return SpacesResponse{
		Total:     s.totalSpaces,
		Occupied:  occupied,
		Available: s.totalSpaces - occupied,
	}
}

func (s *Store) GetDailyIncome(date time.Time) IncomeResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dateStr := date.Format("2006-01-02")
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.Add(24 * time.Hour)

	var records []IncomeRecord
	var totalFee float64

	for _, r := range s.records {
		if r.ExitTime != nil && !r.ExitTime.Before(start) && r.ExitTime.Before(end) {
			records = append(records, IncomeRecord{
				Plate:       r.Plate,
				VehicleType: r.VehicleType,
				EntryTime:   r.EntryTime,
				ExitTime:    *r.ExitTime,
				Fee:         r.Fee,
			})
			totalFee += r.Fee
		}
	}

	if records == nil {
		records = []IncomeRecord{}
	}

	return IncomeResponse{
		Date:     dateStr,
		TotalFee: totalFee,
		Count:    len(records),
		Records:  records,
	}
}

func (s *Store) GetActiveVehicles() []ParkingRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]ParkingRecord, 0, len(s.activeRecords))
	for _, r := range s.activeRecords {
		result = append(result, *r)
	}
	return result
}

func (s *Store) AddMonthly(plate string, durationDays int) (*MonthlySubscription, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	var start, end time.Time

	if existing, ok := s.monthlySubs[plate]; ok && existing.EndDate.After(now) {
		start = existing.EndDate
		end = existing.EndDate.AddDate(0, 0, durationDays)
	} else {
		start = now
		end = now.AddDate(0, 0, durationDays)
	}

	sub := &MonthlySubscription{
		Plate:     plate,
		StartDate: start,
		EndDate:   end,
	}
	s.monthlySubs[plate] = sub
	return sub, nil
}

func (s *Store) isMonthlyValidLocked(plate string) bool {
	sub, ok := s.monthlySubs[plate]
	if !ok {
		return false
	}
	return sub.EndDate.After(time.Now())
}
