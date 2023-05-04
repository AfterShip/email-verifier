package emailverifier

type smtpAPIVerifier interface {
	// isSupported the specific host is supported check by api
	isSupported(host string) bool
	// check must be called before isSupported == true
	check(domain, username string) (*SMTP, error)
}
