package emailverifier

import (
	"testing"
)

var (
	samples = []struct {
		mail   string
		format bool
	}{
		{mail: "example@domain.com", format: true},
		{mail: "support@yahoo.com", format: true},
		{mail: " jerry@gmail.com", format: false},
		{mail: "tool@163.com", format: true},
		{mail: "😀@gmail.com", format: false},
		{mail: "user@gma3il.com", format: true},
		{mail: "a_b@github.com", format: true},
		{mail: "abc@доменное.com", format: true},
	}
)

func TestCheckAddressSyntax(t *testing.T) {
	for _, s := range samples {
		address := verifier.ParseAddress(s.mail)
		if !address.Valid && s.format == true {
			t.Errorf(`"%s" check failed with an unexpected error`, s.mail)
		}
		if address.Valid && s.format == false {
			t.Errorf(`"%s" => incorrect email address`, s.mail)
		}
	}
}
