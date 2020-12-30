package emailverifier

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// Gravatar is detail about the Gravatar
type Gravatar struct {
	HasGravatar bool   // whether has gravatar
	GravatarUrl string // gravatar url
}

// CheckMX will return the Gravatar records for the given email.
func (v *Verifier) CheckGravatar(email string) (*Gravatar, error) {
	emailMd5 := md5V(strings.ToLower(strings.TrimSpace(email)))
	gravatarUrl := gravatarBaseUrl + emailMd5
	resp, err := http.Get(gravatarUrl)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	// check body
	if md5V(string(body)) == gravatarDefaultMd5 || resp.StatusCode != 200 {
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
