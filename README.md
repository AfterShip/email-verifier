# email-verifier

✉️ A Go library for email verification without sending any emails.

[![Build Status](https://github.com/AfterShip/email-verifier/workflows/CI%20Actions/badge.svg)](https://github.com/AfterShip/email-verifier/actions)
[![Godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/aftership/email-verifier)
[![Coverage Status](https://coveralls.io/repos/github/AfterShip/email-verifier/badge.svg?branch=master&t=VTgVfL)](https://coveralls.io/github/AfterShip/email-verifier?branch=master)
[![Go Report](https://goreportcard.com/badge/github.com/aftership/email-verifier)](https://goreportcard.com/report/github.com/aftership/email-verifier)
[![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://github.com/AfterShip/email-verifier/blob/master/LICENSE)

## Features

- Email Address Validation: validates if a string contains a valid email.
- Email Verification Lookup via SMTP: performs an email verification on the passed email
- MX Validation: checks the DNS MX records for the given domain name
- Misc Validation: including Free email provider check, Role account validation, Disposable emails address (DEA) validation
- Email Reachability: checks how confident in sending an email to the address

## Install

Use `go get` to install this package.

```shell script
go get -u github.com/AfterShip/email-verifier
```

## Usage

### Basic usage

Use `Verify` method to verify an email addres with different dimensions

```go
package main

import (
    "fmt"

    "github.com/AfterShip/email-verifier"
)

var (
    verifier = emailverifier.NewVerifier()
)

func main() {

    email := "example@exampledomain.org"
    ret, err := verifier.Verify(email)
    if err != nil {
        fmt.Println("check email failed: ", err)
        return
    }

    fmt.Println("email validation result", ret)
}
```

### Email verification Lookup

Use `CheckSMTP` to performs an email verification lookup via SMTP.

```go
var (
    verifier = emailverifier.
        NewVerifier().
        EnableSMTPCheck()
)

func main() {

    domain := "domain.org"
    ret, err := verifier.CheckSMTP(domain)
    if err != nil {
        fmt.Println("check smtp failed: ", err)
        return
    }

    fmt.Println("smtp validation result: ", ret)

}
```

> Note: because most of the ISPs block outgoing SMTP requests through port 25 to prevent email spamming, the module will not perform SMTP checking by default. You can initialize the verifier with  `EnableSMTPCheck()`  to enable such capability if the port 25 is usable.

### Misc Validation

To check if an email domain is disposable via `IsDisposable`

```go
var (
    verifier = emailverifier.
        NewVerifier().
        EnableAutoUpdateDisposable()
)

func main() {
    domain := "domain.org"
    ret := verifier.IsDisposable(domain)
    fmt.Println("misc validation result: ", ret)
}
```

> Note: It is possible to automatically updating the disposable domains daily by initializing verifier with `EnableAutoUpdateDisposable()`

For more detailed documentation, please check on godoc.org 👉 [email-verifier](https://godoc.org/github.com/aftership/email-verifier)

## Similar Libraries Comparison

|                                     | [email-verifier](https://github.com/AfterShip/email-verifier) | [trumail](https://github.com/trumail/trumail) | [check-if-email-exists](https://reacher.email/) | [freemail](https://github.com/willwhite/freemail) |
| ----------------------------------- | :----------------------------------------------------------: | :-------------------------------------------: | :---------------------------------------------: | :-----------------------------------------------: |
| **Features**                        |                              〰️                              |                      〰️                       |                       〰️                        |                        〰️                         |
| Disposable email address validation |                              ✅                               |       ✅, but not available in free lib        |                        ✅                        |                         ✅                         |
| Disposable address autoupdate       |                              ✅                               |                       🤔                       |                        ❌                        |                         ❌                         |
| Free email provider check           |                              ✅                               |       ✅, but not available in free lib        |                        ❌                        |                         ✅                         |
| Role account validation             |                              ✅                               |                       ❌                       |                        ✅                        |                         ❌                         |
| Syntax validation                   |                              ✅                               |                       ✅                       |                        ✅                        |                         ❌                         |
| Email reachability                  |                              ✅                               |                       ✅                       |                        ✅                        |                         ❌                         |
| DNS records validation              |                              ✅                               |                       ✅                       |                        ✅                        |                         ❌                         |
| Email deliverability                |                              ✅                               |                       ✅                       |                        ✅                        |                         ❌                         |
| Mailbox disabled                    |                              ✅                               |                       ✅                       |                        ✅                        |                         ❌                         |
| Full inbox                          |                              ✅                               |                       ✅                       |                        ✅                        |                         ❌                         |
| Host exists                         |                              ✅                               |                       ✅                       |                        ✅                        |                         ❌                         |
| Catch-all                           |                              ✅                               |                       ✅                       |                        ✅                        |                         ❌                         |
| Gravatar                            |                              🔜                               |       ✅, but not available in free lib        |                        ❌                        |                         ❌                         |
| Typo check                          |                              🔜                               |       ✅, but not available in free lib        |                        ❌                        |                         ❌                         |
| Honeyport dection                   |                              🔜                               |                       ❌                       |                        ❌                        |                         ❌                         |
| Bounce email check                  |                              🔜                               |                       ❌                       |                        ❌                        |                         ❌                         |
| **Tech**                            |                              〰️                              |                      〰️                       |                       〰️                        |                        〰️                         |
| Provide API                         |                              🔜                               |                       ✅                       |                        ✅                        |                         ❌                         |
| Free API                            |                              🔜                               |                       ❌                       |                        ❌                        |                         ❌                         |
| Language                            |                              Go                              |                      Go                       |                      Rust                       |                       Node                        |
| Active maintain                     |                              ✅                               |                       ❌                       |                        ✅                        |                         ✅                         |
| High Performance                   |                              ✅                               |                       ❌                       |                        ✅                        |                         ✅                         |



## FAQ

#### The library hangs/takes a long time after 30 seconds when perform email verification lookup via SMTP

Most ISPs block outgoing SMTP requests through port 25 to prevent email spamming. `email-verifier` needs to have this port open to make a connection to the email's SMTP server. With the port being blocked, it is not possible to perform such checking, and it will instead hang until timeout error. Unfortunately, there is no easy workaround for this issue.

For more information, you may also visit [this StackOverflow thread](https://stackoverflow.com/questions/18139102/how-to-get-around-an-isp-block-on-port-25-for-smtp).

#### The output shows `"connection refused"` in the `smtp.error` field.

This error can also be due to SMTP ports being blocked by the ISP, see the above answer.

#### What does reachable: "unknown" means

This means that the server does not allow real-time verification of an email right now, or the email provider is a catch-all email server.

## Credits

- [trumail](https://github.com/trumail/trumail)
- [check-if-email-exists](https://github.com/amaurymartiny/check-if-email-exists)
- disposable domains from [disposable/disposable](https://github.com/disposable/disposable)
- free provider data from [willwhite/freemail](https://github.com/willwhite/freemail)

## Contributing

For details on contributing to this repository, see the [contributing guide](https://github.com/AfterShip/email-verifier/blob/master/CONTRIBUTING.md).

## License

This package is licensed under MIT license. See [LICENSE](https://github.com/AfterShip/email-verifier/blob/master/LICENSE) for details.
