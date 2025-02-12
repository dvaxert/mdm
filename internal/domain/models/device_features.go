package models

import "github.com/google/uuid"

const (
	Camera  = "camera"
	Storage = "storage"
)

var DefaultFeatures = map[string]bool{
	Camera:  false,
	Storage: false,
}

type DeviceFeatures struct {
	DeviceId   int64
	DeviceUuid uuid.UUID
	Features   map[string]bool
}
