package k8s

import (
	"fmt"
	"time"
)

const defaultCommandTimeout = 5 * time.Minute

func resolveCommandTimeout(timeout time.Duration) (time.Duration, error) {
	if timeout == 0 {
		return defaultCommandTimeout, nil
	}

	if timeout < 0 {
		return 0, fmt.Errorf("timeout cannot be negative")
	}

	return timeout, nil
}
