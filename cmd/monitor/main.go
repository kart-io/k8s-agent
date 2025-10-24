package main

import (
	// Import the automaxprocs package, which automatically configures the GOMAXPROCS
	// value at program startup to match the Linux container's CPU quota.
	_ "go.uber.org/automaxprocs/maxprocs"

	"github.com/kart-io/k8s-agent/cmd/monitor/app"
)

func main() {
	app.Execute()
}
