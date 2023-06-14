package emailverifier

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	glxuPageFormat = "https://mail.google.com/mail/gxlu?email=%s"
)

// See the link below to know why we can use this way to check if a gmail exists.
// https://blog.0day.rocks/abusing-gmail-to-get-previously-unlisted-e-mail-addresses-41544b62b2
func newGmailAPIVerifier(client *http.Client) smtpAPIVerifier {
	if client == nil {
		client = http.DefaultClient
	}
	return gmail{
		client: client,
	}
}

type gmail struct {
	client *http.Client
}

func (g gmail) isSupported(host string) bool {
	return strings.HasSuffix(host, ".google.com.")
}

func (g gmail) check(domain, username string) (*SMTP, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	email := fmt.Sprintf("%s@%s", username, domain)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(glxuPageFormat, email), nil)
	if err != nil {
		return nil, err
	}
	resp, err := g.client.Do(request)
	if err != nil {
		return &SMTP{}, err
	}
	defer resp.Body.Close()
	emailExists := len(resp.Cookies()) > 0

	return &SMTP{
		HostExists:  true,
		Deliverable: emailExists,
	}, nil
}
