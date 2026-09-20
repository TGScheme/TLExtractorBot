package db

import (
	"fmt"
	"time"
)

const (
	waitAttempts = 30
	waitInterval = 2 * time.Second
)

func waitFor(name string, attempt func() error) error {
	var err error
	for range waitAttempts {
		if err = attempt(); err == nil {
			return nil
		}
		time.Sleep(waitInterval)
	}
	return fmt.Errorf("connect to %s: %w", name, err)
}
