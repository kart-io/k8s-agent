package main

import (
	// Import the automaxprocs package, which automatically configures the GOMAXPROCS
	// value at program startup to match the Linux container's CPU quota.
	// This avoids performance issues caused by an inappropriate default GOMAXPROCS
	// value when running in containers, ensuring that the Go program can fully utilize
	// available CPU resources and avoid CPU waste.
	_ "go.uber.org/automaxprocs/maxprocs"

	"github.com/kart-io/k8s-agent/cmd/agent-manager/app"
)

func main() {
	app.Execute()
}
