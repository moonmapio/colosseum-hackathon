package main

import (
	"moonmap.io/go-commons/system"
	"moonmap.io/project-service/service"
)

func main() {

	sys := system.New()

	sys.LoadEnvFile()
	sys.SetFormatter()

	defer sys.Shutdown()

	srv := service.New()
	srv.Start(sys)

}
