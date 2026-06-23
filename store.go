package main

import (
	"sync"
	"time"
)

type Store struct {
	mu            sync.RWMutex
	records       map[string]*ParkingRecord
	parkedVehicles map[string]*ParkingRecord
	parkingLot    *ParkingLot
}

func NewStore(totalSpaces int) *Store {
	return &Store{
		records:        make(map[string]*ParkingRecord),
		parkedVehicles: make(map[string]*ParkingRecord),
		parkingLot: &ParkingLot{
			TotalSpaces: totalSpaces,
			UsedSpaces:  0,
		},
	}
}

func (s *Store) Entry(plateNumber string, vehicleType VehicleType) (*ParkingRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.parkedVehicles[plateNumber]; exists {
		return nil, ErrVehicleAlreadyParked
	}

	if s.parkingLot.UsedSpaces >= s.parkingLot.TotalSpaces {
		return nil, ErrNoAvailableSpaces
	}

	id := generateID()
	record := &ParkingRecord{
		ID:          id,
		PlateNumber: plateNumber,
		VehicleType: vehicleType,
		EntryTime:   time.Now(),
		IsParked:    true,
	}

	s.records[id] = record
	s.parkedVehicles[plateNumber] = record
	s.parkingLot.UsedSpaces++

	return record, nil
}

func (s *Store) Exit(plateNumber string) (*ParkingRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, exists := s.parkedVehicles[plateNumber]
	if !exists {
		return nil, ErrVehicleNotFound
	}

	now := time.Now()
	record.ExitTime = &now
	record.IsParked = false
	record.Fee = calculateFee(record.VehicleType, record.EntryTime, now)

	delete(s.parkedVehicles, plateNumber)
	s.parkingLot.UsedSpaces--

	return record, nil
}

func (s *Store) GetParkingLotStatus() (int, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.parkingLot.TotalSpaces, s.parkingLot.UsedSpaces
}

func (s *Store) GetParkedVehicles() []ParkingRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	vehicles := make([]ParkingRecord, 0, len(s.parkedVehicles))
	for _, r := range s.parkedVehicles {
		vehicles = append(vehicles, *r)
	}
	return vehicles
}

func (s *Store) GetDailyRevenue(date string) *DailyRevenue {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var targetDate time.Time
	if date == "" {
		targetDate = time.Now()
	} else {
		var err error
		targetDate, err = time.Parse("2006-01-02", date)
		if err != nil {
			return &DailyRevenue{Date: date, Records: []ParkingRecord{}, Total: 0}
		}
	}

	year, month, day := targetDate.Date()
	records := []ParkingRecord{}
	total := 0.0

	for _, r := range s.records {
		if r.ExitTime != nil {
			ry, rm, rd := r.ExitTime.Date()
			if ry == year && rm == month && rd == day {
				records = append(records, *r)
				total += r.Fee
			}
		}
	}

	return &DailyRevenue{
		Date:    targetDate.Format("2006-01-02"),
		Records: records,
		Total:   total,
	}
}

func (s *Store) GetVehicleByPlate(plateNumber string) (*ParkingRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	record, exists := s.parkedVehicles[plateNumber]
	if exists {
		return record, true
	}
	return nil, false
}
