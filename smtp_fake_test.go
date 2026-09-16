package emailverifier

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"testing"
	"time"
)

// fakeSMTP answers RCPT for the 32-character random local part GenerateRandomEmail
// produces with randReply, and anything else with userReply, so the catch-all probe
// and the address probe can be driven independently. Returns the listener address.
func fakeSMTP(t *testing.T, randReply, userReply string) string {
	t.Helper()
	var lc net.ListenConfig
	ln, err := lc.Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go serveFakeSMTP(conn, randReply, userReply)
		}
	}()
	return ln.Addr().String()
}

func serveFakeSMTP(conn net.Conn, randReply, userReply string) {
	defer conn.Close()
	br := bufio.NewReader(conn)
	fmt.Fprint(conn, "220 fake ESMTP ready\r\n")
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			return
		}
		command := strings.ToUpper(strings.TrimSpace(line))
		switch {
		case strings.HasPrefix(command, "EHLO"):
			fmt.Fprint(conn, "250-fake\r\n250 OK\r\n")
		case strings.HasPrefix(command, "RCPT"):
			_, path, ok := strings.Cut(command, "<")
			if !ok {
				fmt.Fprint(conn, "501 5.5.4 Syntax error in parameters\r\n")
				continue
			}
			local, _, _ := strings.Cut(path, "@")
			if len(local) == 32 {
				fmt.Fprintf(conn, "%s\r\n", randReply)
			} else {
				fmt.Fprintf(conn, "%s\r\n", userReply)
			}
		case strings.HasPrefix(command, "QUIT"):
			fmt.Fprint(conn, "221 Bye\r\n")
			return
		default:
			fmt.Fprint(conn, "250 OK\r\n")
		}
	}
}

// verifierAgainst returns a Verifier whose SMTP sessions all go to addr.
func verifierAgainst(addr string) *Verifier {
	v := NewVerifier().EnableSMTPCheck()
	v.newClient = func(_, _ string, _ *net.Resolver, _, _ time.Duration) (*smtp.Client, *net.MX, error) {
		var d net.Dialer
		conn, err := d.DialContext(context.Background(), "tcp", addr)
		if err != nil {
			return nil, nil, err
		}
		client, err := smtp.NewClient(conn, "fake.example")
		return client, &net.MX{Host: "fake.example"}, err
	}
	return v
}
