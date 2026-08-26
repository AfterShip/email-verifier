## [Change log](https://github.com/AfterShip/email-verifier/releases)

Unreleased
----------
* Feature: Support a custom DNS resolver for MX and SMTP host lookups via `Resolver()` [#191](https://github.com/AfterShip/email-verifier/pull/191)
* **Breaking**: `LookupError` wraps the error it was derived from, reachable via `errors.Is`/`errors.As`. `ParseSMTPError` no longer returns a nil `*LookupError` for a non-nil input. Adds an unexported field, so whole-struct comparison against a `LookupError` literal no longer matches [#202](https://github.com/AfterShip/email-verifier/pull/202)
* **Breaking**: Remove the non-functional Yahoo API verifier; `EnableAPIVerifier(YAHOO)` and the `YAHOO` constant are gone [#198](https://github.com/AfterShip/email-verifier/pull/198)
* Fix: Yahoo API test panic no longer aborts the test suite [#196](https://github.com/AfterShip/email-verifier/pull/196)

v1.4.0
----------
* Feature: Support Gmail&Yahoo SMTP check by API [#88](https://github.com/AfterShip/email-verifier/pull/88)
* Optimization: Return HasMXRecord as true when at least one valid mx records exist [#94](https://github.com/AfterShip/email-verifier/pull/94)
* Update Dependencies

v1.3.3
----------
* Making catchAll detection optional [#76](https://github.com/AfterShip/email-verifier/pull/76)
* When the user enables `EnableAutoUpdateDisposable()`, the disposable domains configuration is updated once by default.
* Update test cases
* Update Dependencies

v1.3.2
----------
* Uses x/net/proxy to fix issue when using SOCKS5

v1.3.1
----------
* Fix a bug: `IsDisposable()` matches the complete email domain
* Update dependent metadata
* Update Dependencies

v1.3.0
----------
* Support setting SOCKS5 proxy to perform `CheckSMTP()`
* Make pkg compatible with earlier versions of Go

v1.2.0
----------
* Support adding custom disposable email domains 
* Fix a wrong reference in README 
* Update dependent metadata  
* Update Dependencies

v1.1.0
----------
* Performance optimization:
    * reduce Result struct size from 96 to 80
    * `ParseAddress()` return `Syntax` instead of reference, for reducing GC pressure and improve memory locality.
* Provide a simple API server
* Bugfix: gravatar images may not exist

v1.0.3
----------
* Add a New feature: domain suggestion (typo check)

v1.0.2
----------
* Add build metadata tools to generate metadata_*.go files 
* Update load meta data logic
