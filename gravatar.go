package emailverifier

import (
	"io/ioutil"
	"net/http"
	"strings"
)

// Gravatar is detail about the Gravatar
type Gravatar struct {
	HasGravatar bool   // whether has gravatar
	GravatarUrl string // gravatar url
}

// CheckGravatar will return the Gravatar records for the given email.
func (v *Verifier) CheckGravatar(email string) (*Gravatar, error) {
	emailMd5 := getMD5Hash(strings.ToLower(strings.TrimSpace(email)))
	gravatarUrl := gravatarBaseUrl + emailMd5
	resp, err := http.Get(gravatarUrl)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// check body
	if getMD5Hash(string(body)) == gravatarDefaultMd5 || resp.StatusCode != 200 {
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
