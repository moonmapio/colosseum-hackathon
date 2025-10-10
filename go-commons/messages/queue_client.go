package messages

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"moonmap.io/go-commons/helpers"
)

type AlertServiceClient struct {
	BaseURL     string
	ServiceName string
	httpClient  *http.Client
	ctx         context.Context
}

func NewAlertServiceClient(ctx context.Context, serviceName string) *AlertServiceClient {
	baseURL := helpers.GetEnvOrFail("ALERT_SERVICE_URL")
	return &AlertServiceClient{
		ctx:         ctx,
		BaseURL:     baseURL,
		ServiceName: serviceName,
		httpClient:  &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *AlertServiceClient) sendEvent(ctx context.Context, level, message string) {
	hostname, _ := os.Hostname()
	ev := EventMessage{
		ServiceName: c.ServiceName,
		PodName:     hostname,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Message:     message,
		Level:       level,
	}

	body, _ := json.Marshal(ev)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/alerts", c.BaseURL), bytes.NewReader(body))
	if err != nil {
		logrus.Error(err.Error())
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logrus.Error(err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		err := fmt.Errorf("alert service returned %d", resp.StatusCode)
		logrus.Error(err.Error())
		return
	}

}

func (c *AlertServiceClient) EnqueueInfo(message string) {
	c.sendEvent(c.ctx, "info", message)
}

func (c *AlertServiceClient) EnqueueWarn(message string) {
	c.sendEvent(c.ctx, "warn", message)
}

func (c *AlertServiceClient) EnqueueError(message string) {
	c.sendEvent(c.ctx, "error", message)
}
