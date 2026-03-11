package k8s

import "time"

var (
	configFilePath   string
	context          string
	timeout          time.Duration
	parallel         int
	populateExamples bool
	deleteForce      bool
	cleanForce       bool
)
