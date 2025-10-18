package helpers

import (
	"fmt"
	"net/url"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
)

func RewriteForClient(ps *v4.PresignedHTTPRequest) (clientBase, uploadURL string) {

	uploadURL = ps.URL
	u, _ := url.Parse(ps.URL)
	rewriteHost := GetEnv("S3_ENDPOINT_REWRITE", "")
	if rewriteHost != "" {
		ru, _ := url.Parse(rewriteHost)
		u.Scheme = ru.Scheme
		u.Host = ru.Host
		uploadURL = u.String()
	}

	clientBase = fmt.Sprintf("%v://%v/", u.Scheme, u.Host)

	return
}
