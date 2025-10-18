package main

import (
	"moonmap.io/alert-service/service"
	"moonmap.io/go-commons/system"
)

func main() {

	sys := system.New()
	sys.LoadEnvFile()
	sys.SetFormatter()

	defer sys.Shutdown()

	srv := service.New()
	srv.Start(sys)
}
