package emailverifier

import "time"

// New creates a new email verifier, modified by the given options.
// Use "emailverifier.With*" to modify the default behavior.
func New(options ...Option) *Verifier {
	v := NewVerifier()

	for _, option := range options {
		option(v)
	}
	return v
}

// Option represents a modification to the default behavior of a email verifier.
type Option func(*Verifier)

// WithEnableSMTPCheck enables the SMTP check.
func WithEnableSMTPCheck() Option {
	return func(v *Verifier) {
		v.smtpCheckEnabled = true
	}
}

// WithDisableCatchAllCheck enables the catch-all email check.
func WithDisableCatchAllCheck() Option {
	return func(v *Verifier) {
		v.catchAllCheckEnabled = false
	}
}

// WithEnableDomainSuggest enables the domain suggestion if the domain may be misspelled.
func WithEnableDomainSuggest() Option {
	return func(v *Verifier) {
		v.domainSuggestEnabled = true
	}
}

// WithEnableGravatarCheck enables the email check by Gravatar service.
func WithEnableGravatarCheck() Option {
	return func(v *Verifier) {
		v.gravatarCheckEnabled = true
	}
}

// WithFromEmail sets the email address to use as the sender.
func WithFromEmail(email string) Option {
	return func(v *Verifier) {
		v.fromEmail = email
	}
}

// WithHelloName sets the name to use in the `EHLO:` SMTP command.
func WithHelloName(email string) Option {
	return func(v *Verifier) {
		v.helloName = email
	}
}

// WithProxyURI sets verify the email request through specified proxy server.
func WithProxyURI(uri string) Option {
	return func(v *Verifier) {
		v.proxyURI = uri
	}
}

// WithConnectTimeout overrides the default timeout for establishing connections.
func WithConnectTimeout(timeout time.Duration) Option {
	return func(v *Verifier) {
		v.connectTimeout = timeout
	}
}

// WithOperationTimeout overrides the default timeout for each SMTP operation.
func WithOperationTimeout(timeout time.Duration) Option {
	return func(v *Verifier) {
		v.operationTimeout = timeout
	}
}
