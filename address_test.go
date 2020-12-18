package emailverifier

import (
	"testing"
)

var (
	samples = []struct {
		mail   string
		format bool
	}{
		{mail: "florian@carrere.cc", format: true},
		{mail: "support@g2mail.com", format: true},
		{mail: " florian@carrere.cc", format: false},
		{mail: "florian@carrere.cc ", format: false},
		{mail: "test@912-wrong-domain902.com", format: true},
		{mail: "0932910-qsdcqozuioqkdmqpeidj8793@gmail.com", format: true},
		{mail: "@gmail.com", format: false},
		{mail: "test@gmail@gmail.com", format: false},
		{mail: "test test@gmail.com", format: false},
		{mail: " test@gmail.com", format: false},
		{mail: "test@wrong domain.com", format: false},
		{mail: "admin@busyboo.com", format: true},
		{mail: "admin.@busyboo.com", format: false},
		{mail: "abc@中国.com", format: true},
		{mail: "a@gmail.fi", format: true},
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
