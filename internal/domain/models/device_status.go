package models

import "github.com/google/uuid"

type DeviceStatus struct {
	DeviceId   int64
	DeviceUuid uuid.UUID
	Location   string
	Battery    int
}
