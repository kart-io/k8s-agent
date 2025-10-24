package main

import (
	_ "go.uber.org/automaxprocs/maxprocs"

	"github.com/kart-io/k8s-agent/cmd/cluster/app"
)

func main() {
	app.Execute()
}
