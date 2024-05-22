package emailverifier

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// Gravatar is detail about the Gravatar
type Gravatar struct {
	HasGravatar bool   // whether has gravatar
	GravatarUrl string // gravatar url
}

// CheckGravatar will return the Gravatar records for the given email.
func (v *Verifier) CheckGravatar(email string) (*Gravatar, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	emailMd5, err := getMD5Hash(strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		return nil, err
	}
	gravatarUrl := gravatarBaseUrl + emailMd5 + "?d=404"
	req, err := http.NewRequest("GET", gravatarUrl, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// check body
	md5Body, err := getMD5Hash(string(body))
	if err != nil {
		return nil, err
	}
	if md5Body == gravatarDefaultMd5 || resp.StatusCode != 200 {
		return &Gravatar{
			HasGravatar: false,
			GravatarUrl: "",
		}, nil
	}
	return &Gravatar{
		HasGravatar: true,
		GravatarUrl: gravatarUrl,
	}, nil
}
