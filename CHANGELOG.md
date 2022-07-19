# Changelog
All notable changes to this project will be documented in this file.

## [0.0.3] - tbd
- added LDAP support

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
