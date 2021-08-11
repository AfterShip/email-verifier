## Change log

v1.2.0
----------
* Support adding custom disposable email domains https://github.com/AfterShip/email-verifier/pull/31
* Fix a wrong reference in README https://github.com/AfterShip/email-verifier/pull/36
* Update dependent metadata  https://github.com/AfterShip/email-verifier/pull/38 https://github.com/AfterShip/email-verifier/pull/35
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
