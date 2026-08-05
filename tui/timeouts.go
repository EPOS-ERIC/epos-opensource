package tui

import "time"

func defaultK8sTimeout() time.Duration {
	return 10 * time.Minute
}

func defaultK8sTimeoutOption() string {
	return "10m"
}

func k8sTimeoutOptions() []string {
	return []string{"1m", "5m", "10m", "30m", "1h"}
}
