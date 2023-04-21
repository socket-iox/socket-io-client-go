package transport

import (
	"time"
)

const (
	ChannelBufferSize = 100
	OpenMaxRetry      = 10
	OpenPacket        = "0"
	PingPacket        = "2"
	PongPacket        = "3"
	OpenNamespace     = "40"
	NewMessage        = "42"
)

type TransportParams struct {
	Sid          string        `json:"sid"`
	Upgrades     []string      `json:"upgrades"`
	PingInterval time.Duration `json:"pingInterval"`
	PingTimeOut  time.Duration `json:"pingTimeout"`
}

type Stream interface {
	Open() error
	Read() (message []string, err error)
	Write(message string) error
	Close()
}
