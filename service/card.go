package service

import (
	"parking-lot/model"
	"sync"
	"time"
)

type CardService struct {
	mu    sync.RWMutex
	cards map[string]*model.MonthlyCard
}

func NewCardService() *CardService {
	return &CardService{
		cards: make(map[string]*model.MonthlyCard),
	}
}

func (s *CardService) RegisterMonthlyCard(plate, ownerName string, months int) (*model.MonthlyCard, error) {
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

func (s *CardService) RenewMonthlyCard(plate string, months int) (*model.MonthlyCard, error) {
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

func (s *CardService) GetMonthlyCard(plate string) (*model.MonthlyCard, error) {
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

func (s *CardService) IsCardValid(plate string) bool {
	card, err := s.GetMonthlyCard(plate)
	if err != nil {
		return false
	}
	return card.Active && !time.Now().After(card.ExpireDate)
}
