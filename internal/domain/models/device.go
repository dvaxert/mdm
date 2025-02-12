package models

import "github.com/google/uuid"

type DeviceType int

const (
	Android DeviceType = iota
	Ios
	Windows

	DeviceTypeCount
)

type Device struct {
	Id   int64
	Uuid uuid.UUID
	Type DeviceType
}

func (t DeviceType) String() string {
	switch t {
	case Android:
		return "Android"
	case Ios:
		return "Ios"
	case Windows:
		return "Windows"
	}

	return "unknown"
}
