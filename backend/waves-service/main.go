package main

import (
	"moonmap.io/go-commons/persistence"
	"moonmap.io/go-commons/system"
	"moonmap.io/waves-service/service"
)

func main() {
	sys := system.New()
	sys.LoadEnvFile()
	sys.SetFormatter()

	sys.AddCleanUpHook(persistence.CloseMongo)
	defer sys.Shutdown()

	srv := service.New()
	srv.Start(sys)
}
