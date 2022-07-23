---
title: API Reference

toc_footers:
  - <a href='https://github.com/burakkavak/nginx-auth-server'>nginx-auth-server on GitHub</a>
  - <a href='https://github.com/slatedocs/slate'>Documentation Powered by Slate</a>

search: true

code_clipboard: true

---

# Command-line interface (CLI)

The nginx-auth-server CLI is used for user and cookie management.

## Usage information

```shell
$ nginx-auth-server help
$ nginx-auth-server --help
```

You can add the *help* argument or the *--help* flag to any command to get usage information for the given command.

## Add user

```shell
$ nginx-auth-server user add --username foo --password foobar --otp
$ nginx-auth-server u a -u foo -p foobar -o
```

This command will add and persist a user in the database. The command will fail if there is an existing user with the same username. If there is existing cookies related to this username (maybe the user was authenticated through LDAP in the past), the command will also fail. If TOTP is required for this user, the application will attempt to print the TOTP secret QR code in the terminal using the [QRencode](https://github.com/fukuchi/libqrencode) library.

### Command flags

Flag           | Meaning
-------------- | ----------
-u, --username | Required. Username of the new user.
-p, --password | Required. Password of the new user.
-o, --otp      | Optional. Require 2FA for this user.

## Remove user

```shell
$ nginx-auth-server user remove --username foo
$ nginx-auth-server u r -u foo
```

This command will remove a user from the database. All user-related cookies will also be deleted in the process.

### Command flags

Flag           | Meaning
-------------- | ----------
-u, --username | Required. Username of the existing user.

## List all users

```shell
$ nginx-auth-server user list
$ nginx-auth-server u l
```

> The above command returns JSON structured like this:

```json
[
  {
    "username": "foo",
    "password": "$2a$04$P/PP2/wTdzwAZJphoWJjE.srUTKPbIA2If19WBmHrZ0E9As.0TilO",
    "otpSecret": null
  }
]
```

This command will list all users in the database.

## List all cookies

```shell
$ nginx-auth-server cookie list
$ nginx-auth-server c l
```

> The above command returns JSON structured like this:

```json
the database contains 1 cookies
[
  {
    "name": "Nginx-Auth-Server-Token",
    "value": "$2a$04$lPwzADvbNB2jkNo5l8uNAO8BFmbe6jvXLG649L3VG7VRtCtuU8GKi",
    "expires": "2022-07-30T22:09:27.1577237+02:00",
    "domain": "localhost",
    "username": "foo",
    "httpOnly": true,
    "secure": true
  }
]
```

This command will list all cookies in the database.

## Purge all cookies

```shell
$ nginx-auth-server cookie purge
$ nginx-auth-server c p
```

This command will delete all cookies in the database. All user authorizations will be reset in the process.

# HTTP API

nginx-auth-server uses cookies to authorize with the API. The cookie is set by the server in the response of a successful login (see [login](#login-form)).

## Authenticate

```javascript
fetch("http://localhost:17397/auth")
```

This endpoint checks if the client is authenticated.

### HTTP Request

`GET http://localhost:17397/auth`

### HTTP status codes

Status code | Meaning
----------- | -------
200         | Client is authenticated and the provided auth cookie is valid.
401         | Client is not authenticated or the auth cookie expired.

## Login (HTML)

```javascript
fetch("http://localhost:17397/login")
```

This endpoint provides the HTML for the login page. If the client is already authenticated and the auth cookie that was provided is valid, the endpoint provides an empty response.

### HTTP Request

`GET http://localhost:17397/login`

## Login form

```javascript
fetch('http://localhost:17397/login', {
  method: 'post',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    inputUsername: 'foo',
    inputPassword: 'bar',
    inputTotp: '123456',
    recaptchaToken: '<google_provided_recaptcha_token_here>'
  }),
});
```

This endpoint processes the information from the login form and sets the authentication cookie if the client provides valid credentials.

### HTTP Request

`POST http://localhost:17397/login`

### POST body (encoded as JSON)

Parameter      | Description
---------      | -----------
inputUsername  | Required. The username provided by the user.
inputPassword  | Required. The password provided by the user.
inputTotp      | Required if TOTP is enabled for the provided user. The six digit one time password.
recaptchaToken | Required if reCAPTCHA is enabled in the server configuration. The reCAPTCHA token provided by Google upon execution.

### HTTP status codes

<aside class="notice">The endpoint may provide additional error information as JSON in the response.</aside>

Status code | Meaning
----------- | -------
200         | Client was already or is freshly authenticated. The auth cookie will be set by the endpoint if the user was not already authenticated.
401         | Bad input: username, password or TOTP is invalid.
500         | Internal server error regarding reCAPTCHA verification.

## Logout

```javascript
fetch("http://localhost:17397/logout")
```

This endpoint will immediately expire the authentication cookie if the client is authenticated.

### HTTP Request

`GET http://localhost:17397/logout`

### HTTP status codes

Status code | Meaning
----------- | -------
200         | Client was successfully logged out.
401         | Client is not authenticated.

## Whoami

```javascript
fetch("http://localhost:17397/whoami")
```

> The above command returns JSON structured like this:

```json
{
  "username": "foo",
}
```

This endpoint will return the clients username if the client is authenticated.

### HTTP Request

`GET http://localhost:17397/whoami`

### HTTP status codes

Status code | Meaning
----------- | -------
200         | User information has been returned in the response.
401         | Client is not authenticated.