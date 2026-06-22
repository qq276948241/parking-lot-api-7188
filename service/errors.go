package service

import "errors"

var (
	ErrVehicleAlreadyIn = errors.New("车辆已在场内")
	ErrNoSpace          = errors.New("车位已满")
	ErrInvalidCarType   = errors.New("无效的车辆类型，可选: temp, monthly")
	ErrVehicleNotFound  = errors.New("该车辆不在场内")
	ErrMonthlyCardExists   = errors.New("该车牌已注册月卡")
	ErrMonthlyCardNotFound = errors.New("该车牌未注册月卡")
	ErrMonthlyCardExpired  = errors.New("月卡已过期，请续费")
	ErrInvalidMonths       = errors.New("续费月数必须大于0")
)
