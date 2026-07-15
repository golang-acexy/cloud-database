package gorm

import (
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	if err := starterLoader.Start(); err != nil {
		os.Exit(1)
	}
	code := m.Run()
	if _, err := starterLoader.StopAllByRegisteredOrder(10 * time.Second); err != nil {
		os.Exit(1)
	}
	os.Exit(code)
}
