## [Change log](https://github.com/AfterShip/email-verifier/releases)

Unreleased
----------
* Internal: Lint the whole repository rather than only changed lines, clearing 42 pre-existing findings; enable `usestdlibvars` and `intrange`; fix six error assertions that only checked that an error was non-nil [#217](https://github.com/AfterShip/email-verifier/pull/217)

v1.5.0
----------
* **Breaking**: Supported Go versions are now the three most recent releases, and `go.mod` requires the oldest of them, Go 1.25. The versions of `golang.org/x/net` and `golang.org/x/text` that fix [GO-2026-5026](https://pkg.go.dev/vuln/GO-2026-5026) and [GO-2026-5970](https://pkg.go.dev/vuln/GO-2026-5970) require it, and both were reachable from `Verify`, `CheckMX` and `CheckSMTP` through `idna.ToASCII` -- the second of them an infinite loop on caller-supplied input [#212](https://github.com/AfterShip/email-verifier/pull/212)
* Fix: `AddDisposableDomains` no longer races the refresh that `EnableAutoUpdateDisposable()` schedules. Calling it from another goroutine while a refresh was running was a `fatal error: concurrent map iteration and map write`, which brings the process down and no `recover` can catch [#211](https://github.com/AfterShip/email-verifier/pull/211)
* Fix: A small allowlist, `cmd/build_metadata/disposable_allowlist.txt`, keeps real providers the upstream list misclassifies out of the disposable set -- currently `hush.com`, `lavabit.com`, `mail2world.com`, `qmail.com`, `4x4man.com` and `alphafrau.de`. It applies to the generated list and to every `EnableAutoUpdateDisposable()` refresh, and since the free list is the free candidates minus the disposable one, these domains are now reported free rather than neither [#209](https://github.com/AfterShip/email-verifier/pull/209)
* **Breaking**: The generated disposable list now comes from the same source `EnableAutoUpdateDisposable()` fetches at runtime. The two had diverged, so enabling auto-update used to replace most of the baked-in list; it is now a no-op. `IsDisposable` gains 37679 domains and loses 96017, the vast majority of which no longer resolve [#208](https://github.com/AfterShip/email-verifier/pull/208)
* Fix: The free-domain list is built from its original sources instead of an aggregate that had begun merging disposable blocklists into itself. `IsFreeDomain` and `SuggestDomain` no longer treat throwaway domains as free providers, 108 carrier and portal mailboxes are recognised, and `IsFreeDomain("atlanticbb.net")` works -- the generated key carried a stray no-break space [#207](https://github.com/AfterShip/email-verifier/pull/207)
* Fix: `IsRoleAccount` recognises `cto`, `ctos`, `cfo` and `cfos`. They were added to the source list in December 2025 but the generated map was never rebuilt, so they returned `false` until now [#205](https://github.com/AfterShip/email-verifier/pull/205)
* Feature: Support a custom DNS resolver for MX and SMTP host lookups via `Resolver()` [#191](https://github.com/AfterShip/email-verifier/pull/191)
* Fix: `Verify` keeps the SMTP result when the exchange fails partway. `SMTP` used to come back nil, which could not distinguish a host that never answered from one that answered and refused the sender [#201](https://github.com/AfterShip/email-verifier/pull/201)
* Fix: `Reachable` is `no`, not `unknown`, when the MX lookup fails with `no such host`, and `Verify` returns `ErrNoSuchHost` [#139](https://github.com/AfterShip/email-verifier/pull/139)
* Fix: `Suggestion` is filled in for disposable domains, which used to return early before the suggestion was computed [#140](https://github.com/AfterShip/email-verifier/pull/140)
* **Breaking**: The free-domain sources drop six forks of one 2014 gist for the list they were forked from. One of the six was contributing throwaway domains rather than providers -- 77% of its 23,882 entries appear on disposable blocklists -- so `IsFreeDomain` loses 2,009 domains it should not have had [#169](https://github.com/AfterShip/email-verifier/pull/169)
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
