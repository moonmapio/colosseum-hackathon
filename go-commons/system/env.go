package system

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"moonmap.io/go-commons/helpers"
)

func (s *System) LoadEnvFile() {

	environment := helpers.GetEnv("ENVIRONMENT", "dev")

	if environment == "dev" {
		cwd, _ := os.Getwd()
		path := fmt.Sprintf("%v/.env", cwd)

		if err := godotenv.Load(path); err != nil {
			logrus.WithError(err).Warn("ℹ️  .env not found or couldnt be loaded")
		} else {
			logrus.WithField("paths", path).Info("🔑 .env load")
		}
	}

	s.Bind = helpers.GetEnv("BIND", ":8080")
	if !strings.HasPrefix(s.Bind, ":") {
		s.Bind = ":" + s.Bind
	}

}
