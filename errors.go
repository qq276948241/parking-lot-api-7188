package main

import "errors"

var (
	ErrVehicleAlreadyParked = errors.New("车辆已在场内")
	ErrNoAvailableSpaces = errors.New("无可用车位")
	ErrVehicleNotFound    = errors.New("车辆未找到")
	ErrInvalidVehicleType = errors.New("无效的车辆类型")
	ErrInvalidPlateNumber = errors.New("车牌号不能为空")
)
