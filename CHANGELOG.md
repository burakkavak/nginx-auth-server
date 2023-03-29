# Changelog
All notable changes to this project will be documented in this file.

## [0.0.10] - TBD
- added support for changing the cookie name (see: https://github.com/burakkavak/nginx-auth-server/issues/3)

## [0.0.9] - 2023-03-23
- fixed IP address logging upon authentication to log the real client IP
- fixed JS/CSS caching problems with browsers
- */login* route now redirects to a callback URL instead of returning a 200 status code when the user is already authenticated
- updated dependencies (including an update to TypeScript 5 and Golang 1.20.x). this also should fix some security vulnerabilities with *golang.org/x/net*

## [0.0.8] - 2023-01-15
- performance optimization to mitigate argon2 performance hit
- added 'show password' functionality in the login form
- improved logging: application logs to 'auth.log' and 'app.log' now
- fixed long Docker container shutdown time
- fixed bug where any TOTP input was rejected

## [0.0.7] - 2023-01-11
- Docker support
- support for HTTPS/TLS
- password parameter removed from user creation to avoid bash/zsh entries with plain text passwords
- expired cookies are automatically deleted from the database and cache now
- check for existing (LDAP) users before user creation

## [0.0.6] - 2023-01-05
- updated security (*argon2* is being used instead of *bcrypt* now)
- implemented cache for better performance
- cookie lifetime and secure attribute is now customizable in config.ini
- upon cookie expiration, the user will get a notice in the frontend now
- updated Go to version 1.19
- updated Go dependencies
- updated Node.js dependencies

## [0.0.5] - 2022-08-25
- fixed an issue with the libc-dependency in the binaries, that prevented the application from running on older libc versions

## [0.0.4] - 2022-07-22
- added /whoami API endpoint

## [0.0.3] - 2022-07-22
- added LDAP support with [go-ldap/ldap](https://github.com/go-ldap/ldap)
- added Google reCAPTCHA v2 support for login form

## [0.0.2] - 2022-07-18
- added TOTP support with [pquerna/otp](https://github.com/pquerna/otp)
- added [QRencode](https://github.com/fukuchi/libqrencode) support to display generated TOTP secrets in terminal
- changed/improved frontend form validation
- remove associated user cookies upon user deletion
- improved cookie security

## [0.0.1] - 2022-07-17
- initial MVP implementation
- implemented frontend login form
- implemented authentication cookie
- implemented data persistence with [etcd-io/bbolt](https://github.com/etcd-io/bbolt)
- implemented configuration parsing with [go-ini/ini](https://github.com/go-ini/ini)
- implemented webserver with [gin-gonic/gin](https://github.com/gin-gonic/gin)
- implemented CLI with [urfave/cli](https://github.com/urfave/cli)
