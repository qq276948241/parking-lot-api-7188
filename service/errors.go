package service

import "errors"

var (
	ErrVehicleAlreadyIn = errors.New("车辆已在场内")
	ErrNoSpace          = errors.New("车位已满")
	ErrInvalidCarType   = errors.New("无效的车辆类型，可选: temp, monthly")
	ErrVehicleNotFound  = errors.New("该车辆不在场内")
)
