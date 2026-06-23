package main

import "time"

const (
	FirstHourFee      = 5.0
	AdditionalHourFee = 3.0
	DailyMaxFee       = 50.0
	FreeMinutes       = 15
)

func CalculateFee(duration time.Duration) float64 {
	totalMin := duration.Minutes()

	if totalMin <= FreeMinutes {
		return 0
	}

	hours := totalMin / 60
	remaining := totalMin - float64(int(hours))*60

	billableHours := int(hours)
	if remaining > 0 {
		billableHours++
	}

	if billableHours == 0 {
		return 0
	}

	var fee float64
	fee = FirstHourFee
	billableHours--

	for billableHours > 0 {
		fee += AdditionalHourFee
		billableHours--
	}

	if fee > DailyMaxFee {
		fee = DailyMaxFee
	}

	return fee
}
