## [Change log](https://github.com/AfterShip/email-verifier/releases)

Unreleased
----------
* Fix: `AddDisposableDomains` no longer races the refresh that `EnableAutoUpdateDisposable()` schedules. Calling it from another goroutine while a refresh was running was a `fatal error: concurrent map iteration and map write`, which brings the process down and no `recover` can catch [#211](https://github.com/AfterShip/email-verifier/pull/211)
* Fix: A small allowlist, `cmd/build_metadata/disposable_allowlist.txt`, keeps real providers the upstream list misclassifies out of the disposable set -- currently `hush.com`, `lavabit.com`, `mail2world.com`, `qmail.com`, `4x4man.com` and `alphafrau.de`. It applies to the generated list and to every `EnableAutoUpdateDisposable()` refresh, and since the free list is the free candidates minus the disposable one, these domains are now reported free rather than neither [#209](https://github.com/AfterShip/email-verifier/pull/209)
* **Breaking**: The generated disposable list now comes from the same source `EnableAutoUpdateDisposable()` fetches at runtime. The two had diverged, so enabling auto-update used to replace most of the baked-in list; it is now a no-op. `IsDisposable` gains 37679 domains and loses 96017, the vast majority of which no longer resolve [#208](https://github.com/AfterShip/email-verifier/pull/208)
* Fix: The free-domain list is built from its original sources instead of an aggregate that had begun merging disposable blocklists into itself. `IsFreeDomain` and `SuggestDomain` no longer treat throwaway domains as free providers, 108 carrier and portal mailboxes are recognised, and `IsFreeDomain("atlanticbb.net")` works -- the generated key carried a stray no-break space [#207](https://github.com/AfterShip/email-verifier/pull/207)
* Fix: `IsRoleAccount` recognises `cto`, `ctos`, `cfo` and `cfos`. They were added to the source list in December 2025 but the generated map was never rebuilt, so they returned `false` until now [#205](https://github.com/AfterShip/email-verifier/pull/205)
* Feature: Support a custom DNS resolver for MX and SMTP host lookups via `Resolver()` [#191](https://github.com/AfterShip/email-verifier/pull/191)
* **Breaking**: `LookupError` wraps the error it was derived from, reachable via `errors.Is`/`errors.As`. `ParseSMTPError` no longer returns a nil `*LookupError` for a non-nil input. Adds an unexported field, so whole-struct comparison against a `LookupError` literal no longer matches [#202](https://github.com/AfterShip/email-verifier/pull/202)
* **Breaking**: Remove the non-functional Yahoo API verifier; `EnableAPIVerifier(YAHOO)` and the `YAHOO` constant are gone [#198](https://github.com/AfterShip/email-verifier/pull/198)
* Fix: Yahoo API test panic no longer aborts the test suite [#196](https://github.com/AfterShip/email-verifier/pull/196)

v1.4.1
----------
* Feature: Configurable timeouts via `ConnectTimeout()` and `OperationTimeout()` [#110](https://github.com/AfterShip/email-verifier/pull/110)
* **Breaking**: Remove the Gmail API verifier; `EnableAPIVerifier(GMAIL)` and the `GMAIL` constant are gone [#113](https://github.com/AfterShip/email-verifier/pull/113)
* Update the free domain sources and rebuild the generated metadata [#127](https://github.com/AfterShip/email-verifier/pull/127)
* Update Dependencies

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
