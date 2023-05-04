package emailverifier

import (
	"fmt"
	"net/http"
	"strings"
)

const (
	glxuPageFormat = "https://mail.google.com/mail/gxlu?email=%s"
)

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
	email := fmt.Sprintf("%s@%s", username, domain)
	resp, err := g.client.Get(fmt.Sprintf(glxuPageFormat, email))
	if err != nil {
		return &SMTP{}, err
	}

	emailExists := len(resp.Cookies()) > 0

	return &SMTP{
		HostExists:  true,
		Deliverable: emailExists,
	}, nil
}
