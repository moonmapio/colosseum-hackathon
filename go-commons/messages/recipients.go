package messages

import (
	"strings"

	"moonmap.io/go-commons/helpers"
)

func GetNotifyRecipients() []*EmailRecipient {
	raw := helpers.GetEnvOrFail("NOTIFY_EMAILS")
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	recipients := make([]*EmailRecipient, 0, len(parts))

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		// Formato esperado: Name:Email
		sub := strings.SplitN(p, ":", 2)
		if len(sub) == 2 {
			recipients = append(recipients, &EmailRecipient{
				Name:  strings.TrimSpace(sub[0]),
				Email: strings.TrimSpace(sub[1]),
			})
		} else {
			// fallback: solo email sin nombre
			recipients = append(recipients, &EmailRecipient{
				Name:  "MoonMap User",
				Email: sub[0],
			})
		}
	}

	return recipients
}
