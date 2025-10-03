package messages

import (
	"context"
	"os"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

type QueueManager struct {
	ServiceName    string
	eventQueue     chan EventMessage
	ctx            context.Context
	recipients     []*EmailRecipient
	mailjetManager *MailjetManager
	startedAt      time.Time
	closed         int32 // atomic flag

}

type EventMessage struct {
	ServiceName string
	PodName     string
	Timestamp   string
	Message     string
	Level       string // "info", "warn", "error"
}

type Data struct {
	Messages []EventMessage
}

func NewManager(serviceName, email, name string, startedAt time.Time) *QueueManager {
	return NewManagerWithSize(serviceName, email, name, startedAt, 1000)
}

func NewManagerWithSize(serviceName, email, name string, startedAt time.Time, channelize int) *QueueManager {
	m := &QueueManager{eventQueue: make(chan EventMessage, channelize)}

	m.startedAt = startedAt
	m.recipients = GetNotifyRecipients()
	m.ServiceName = serviceName

	mjmanager := New(email, name)
	mjmanager.SetTemplate("service_status", m.GetEmailTemplate())
	m.mailjetManager = mjmanager

	return m
}

func (m *QueueManager) Close() {
	atomic.StoreInt32(&m.closed, 1)
	close(m.eventQueue)
}

func (m *QueueManager) SetContext(ctx context.Context) {
	m.ctx = ctx
}

func (m *QueueManager) StartAggregator(freq time.Duration) {
	ticker := time.NewTicker(freq)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-m.ctx.Done():
				m.flushEvents() // opcional: último flush al salir
				return
			case <-ticker.C:
				m.flushEvents()
			}
		}
	}()
}

func (m *QueueManager) flushEvents() {
	var events []EventMessage
	drain := true
	for drain {
		select {
		case ev := <-m.eventQueue:
			events = append(events, ev)
		default:
			drain = false
		}
	}

	if len(events) == 0 {
		return
	}

	data := &Data{Messages: events}
	err := m.mailjetManager.SendMail(
		m.recipients,
		"Cluster Events",
		"See attached report",
		data,
	)

	if err != nil {
		logrus.WithError(err).Error("failed to send aggregated events email")
	} else {
		logrus.Infof("sent aggregated email with %d events", len(events))
	}
}

func (m *QueueManager) EnqueueMessage(serviceName, message string, level string) {
	if atomic.LoadInt32(&m.closed) == 1 {
		return // ignora en colas cerradas lógicamente
	}

	hostname, err := os.Hostname()
	if err != nil {
		logrus.Panic(err)
	}
	now := time.Now().UTC().Format(time.RFC3339)

	msg := EventMessage{
		ServiceName: serviceName,
		PodName:     hostname,
		Timestamp:   now,
		Message:     message,
		Level:       level,
	}

	m.eventQueue <- msg
}

func (m *QueueManager) EnqueueError(serviceName, message string) {
	m.EnqueueMessage(serviceName, message, "error")
}

func (m *QueueManager) EnqueueInfo(serviceName, message string) {
	m.EnqueueMessage(serviceName, message, "info")
}

func (m *QueueManager) EnqueueWarn(serviceName, message string) {
	m.EnqueueMessage(serviceName, message, "warn")
}

func (s *QueueManager) GetEmailTemplate() *string {
	emailHTMLTemplate := `
	<!DOCTYPE html>
	<html lang="en">
		<head>
			<meta charset="UTF-8" />
			<title>Cluster Events Report</title>
			<meta name="color-scheme" content="light" />
			<meta name="supported-color-schemes" content="light" />
			<style>
			body {
				margin: 0;
				padding: 20px;
				background-color: #ffffff;
				color: #111827;
				font-family: "Helvetica Neue", Arial, sans-serif;
				line-height: 1.5;
			}
			h1 {
				font-size: 20px;
				font-weight: 600;
				margin-bottom: 20px;
			}
			table {
				width: 100%;
				border-collapse: collapse;
				margin-top: 10px;
			}
			th, td {
				border: 1px solid #e5e7eb;
				padding: 8px 12px;
				text-align: left;
				font-size: 14px;
			}
			th {
				background-color: #f9fafb;
				font-weight: 600;
			}
			tr:nth-child(even) {
				background-color: #f3f4f6;
			}
			.status-ok {
				color: #059669; /* verde */
				font-weight: bold;
			}
			.status-error {
				color: #dc2626; /* rojo */
				font-weight: bold;
			}
			.status-warn {
				color: #d97706; /* naranja */
				font-weight: bold;
			}
			</style>
		</head>
		<body>
			<h1>Cluster Events Report</h1>
			<table>
			<thead>
				<tr>
					<th>Service</th>
					<th>Pod Name</th>
					<th>Timestamp</th>
					<th>Message</th>
					<th>Status</th>
				</tr>
			</thead>
			<tbody>
				{{ range .Messages }}
				<tr>
					<td>{{ .ServiceName }}</td>
					<td>{{ .PodName }}</td>
					<td>{{ .Timestamp }}</td>
					<td>{{ .Message }}</td>
					<td class="{{ if eq .Level "error" }}status-error{{ else if eq .Level "warn" }}status-warn{{ else }}status-ok{{ end }}">
						{{ if eq .Level "error" }}ERROR{{ else if eq .Level "warn" }}WARN{{ else }}OK{{ end }}
					</td>
				</tr>
				{{ end }}
			</tbody>
			</table>
		</body>
	</html>
	`
	return &emailHTMLTemplate
}
