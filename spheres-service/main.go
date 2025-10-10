package main

import (
	"moonmap.io/go-commons/persistence"
	"moonmap.io/go-commons/system"
	"moonmap.io/spheres-service/core"
)

func main() {
	sys := system.New()
	sys.LoadEnvFile()
	sys.SetFormatter()

	sys.AddCleanUpHook(persistence.CloseMongo)
	defer sys.Shutdown()

	srv := core.NewService()
	srv.Start(sys)
}
