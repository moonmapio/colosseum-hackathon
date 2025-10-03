package helpers

import (
	"errors"
	"net"
	"strings"
)

func DnsErr(err error) *net.DNSError {
	var dnse *net.DNSError
	if errors.As(err, &dnse) {
		return dnse
	}
	// algunos resolvers envuelven distinto: fallback por string match
	if err != nil && strings.Contains(err.Error(), "no such host") {
		return &net.DNSError{Err: "no such host", IsNotFound: true}
	}
	return nil
}
