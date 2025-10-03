package helpers

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// getenv con valor por defecto
func GetEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return strings.TrimSpace(v)
	}
	return def
}

func GetEnvOrFail(k string) string {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		logrus.Fatalf("missing required env var: %s", k)
	}
	return v
}

func GetEnvInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return def
}

func GetEnvDur(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
