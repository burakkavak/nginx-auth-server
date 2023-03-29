# nginx-auth-server

A lightweight authentication server designed to be used in conjunction with nginx 'http_auth_request_module'. nginx-auth-server provides an additional authentication layer that is useful for reverse proxy scenarios, where the proxied service does not support user authentication.

## Table of Contents

- [Demo](#demo)
- [Features](#features)
- [Getting Started](#getting-started)
  - [With Docker](#with-docker)
  - [Native](#native)
- [Contributing](#contributing)
- [Documentation](#documentation)
- [Changelog](#changelog)
- [Credits](#credits)
- [License](#license)

## Demo

![demo.gif](https://raw.githubusercontent.com/burakkavak/nginx-auth-server/master/assets/demo.gif)

## Features

- low latency (<1ms)
- support for Two-Factor Authentication (2FA)
- support for LDAP to validate user credentials
- optional bot protection with Google reCAPTCHA

## Getting Started

### With Docker

Download the [docker-compose.yml](https://raw.githubusercontent.com/burakkavak/nginx-auth-server/master/docker-compose.yml) into a directory of your liking.
```shell
$ mkdir -p ~/docker/nginx-auth-server
$ cd ~/docker/nginx-auth-server
$ wget --content-disposition https://raw.githubusercontent.com/burakkavak/nginx-auth-server/master/docker-compose.yml
```

Start the container:
```shell
$ docker-compose up -d
```

You can now point a NGINX server to this docker container, please refer to the '[Native](#native)' section for a NGINX configuration example.

The CLI is called using *docker exec*. Please refer to the [CLI reference](https://burakkavak.github.io/nginx-auth-server/#command-line-interface-cli) for all commands. Here are some examples:
```shell
# docker exec -it <container_name> nginx-auth-server <command_parameters>
$ docker exec -it nginx-auth-server nginx-auth-server user add --username foo
$ docker exec -it nginx-auth-server nginx-auth-server user list
$ docker exec -it nginx-auth-server nginx-auth-server cookie list
```

#### Environment variables

The docker application can be configured using environment variables. Modify the *docker-compose.yml* and restart the container so the changes take effect.

|    Environment variable    |               Default value               | Description                                                                                                                         |
|:--------------------------:|:-----------------------------------------:|-------------------------------------------------------------------------------------------------------------------------------------|
|  `SERVER_LISTEN_ADDRESS`   |                 `0.0.0.0`                 | The HTTP(S) server is listening to requests on this address (inside the container)                                                  |
|    `SERVER_LISTEN_PORT`    |                  `17397`                  | The application is going to listen for HTTP requests on this port (inside the container)                                            |
|      `SERVER_DOMAIN`       |                `localhost`                | Domain used to set the authentication cookie and as issuer for TOTP. E.g. `example.org`                                             |
|       `TLS_ENABLED`        |                  `false`                  | Enable HTTPS/TLS encryption for the webserver. The unencrypted HTTP server will be disabled                                         |
|     `TLS_LISTEN_PORT`      |                  `17760`                  | The application is going to listen for HTTPS requests on this port (inside the container)                                           |
|      `TLS_CERT_PATH`       | `/opt/nginx-auth-server/certs/server.crt` | Path of the SSL certificate (inside the container)                                                                                  |
|       `TLS_CERT_KEY`       | `/opt/nginx-auth-server/certs/server.key` | Path of the SSL certificate key (inside the container)                                                                              |
|     `COOKIES_LIFETIME`     |                    `7`                    | Cookie lifetime in days. User has to re-authenticate after expiration                                                               |
|      `COOKIES_SECURE`      |                  `true`                   | Set secure attribute for cookies. The browser will only send the auth cookie in a HTTPS context if this is enabled                  |
|       `LDAP_ENABLED`       |                  `false`                  | Enable/disable LDAP support. The application will prioritize local authentication data first                                        |
|         `LDAP_URL`         |                                           | LDAP url. Example for TLS connection: `ldaps://ldap.example.com:636`. Example for non-TLS connection: `ldap://ldap.example.com:389` |
| `LDAP_ORGANIZATIONAL_UNIT` |                  `users`                  | LDAP organizational unit (OU) that is used to search the user                                                                       |
|  `LDAP_DOMAIN_COMPONENTS`  |                                           | LDAP baseDN (DC) of the LDAP tree. Example: `dc=example,dc=org`                                                                     |
|    `RECAPTCHA_ENABLED`     |                  `false`                  | Enable/disable Google reCAPTCHA v2 (invisible) support for the login form                                                           |
|    `RECAPTCHA_SITE_KEY`    |                                           | reCAPTCHA site key that is provided by Google upon site creation                                                                    |
|   `RECAPTCHA_SECRET_KEY`   |                                           | reCAPTCHA secret key that is provided by Google upon site creation                                                                  |

### Native

Download the appropriate binary from the [Releases](https://github.com/burakkavak/nginx-auth-server/releases) section.

Download the current [config.ini](https://raw.githubusercontent.com/burakkavak/nginx-auth-server/master/config.ini) into the same directory:

```shell
$ wget --content-disposition https://raw.githubusercontent.com/burakkavak/nginx-auth-server/master/config.ini
```

Run the server:
```shell
$ ./nginx-auth-server run
```

For user management (adding/removing users) refer to the CLI usage information:
```shell
$ ./nginx-auth-server help
$ ./nginx-auth-server user add --username foo --otp
```

Reconfigure nginx server:
```nginx
server {
  listen 80 default_server;
  listen [::]:80 default_server;

  root /var/www/html;

  index index.html index.htm index.nginx-debian.html;

  server_name _;

  # Redirect user to /login if nginx-auth-server responds with '401 Unauthorized'
  error_page 401 /login;

  location / {
    auth_request /auth;

    # pass Set-Cookie headers from the subrequest response back to requestor
    auth_request_set $auth_cookie $upstream_http_set_cookie;
    add_header Set-Cookie $auth_cookie;

    auth_request_set $auth_status $upstream_status;

    # serve files if the user is authenticated
    try_files $uri $uri/ /index.html;
  }

  location = /auth {
    # internally only, /auth can not be accessed from outside
    internal;

    # nginx-auth-server running on port 17397
    proxy_pass http://localhost:17397;

    # don't pass request body to proxied server, we only need the headers which are passed on by default
    proxy_pass_request_body off;

    # there is no content length since we stripped the request body
    proxy_set_header Content-Length "";

    # let proxy server know more details of request
    proxy_set_header X-Original-URI $request_uri;
    proxy_set_header X-Original-Remote-Addr $remote_addr;
    proxy_set_header X-Original-Host $host;
  }

  # these are handled by nginx-auth-server as part of the auth routines
  location ~ ^/(login|logout|whoami)$ {
    proxy_pass http://localhost:17397;

    proxy_set_header X-Original-URI $request_uri;
    proxy_set_header X-Original-Remote-Addr $remote_addr;
    proxy_set_header X-Original-Host $host;
  }

  # static nginx-auth-server assets (css, js, ...)
  location /nginx-auth-server-static {
    proxy_pass http://localhost:17397/nginx-auth-server-static;

    proxy_set_header X-Original-URI $request_uri;
    proxy_set_header X-Original-Remote-Addr $remote_addr;
    proxy_set_header X-Original-Host $host;
  }
}
```

You can also run the server as a systemd service. Example configuration for user *www-data*:
```apacheconf
[Unit]
Description=nginx-auth-server
After=network.target

[Service]
Type=simple
User=www-data
Group=www-data
WorkingDirectory=/var/www/nginx-auth-server
ExecStart=/var/www/nginx-auth-server/nginx-auth-server run
Restart=on-failure
# Other restart options: always, on-abort, etc

# The install section is needed to use
# `systemctl enable` to start on boot
# For a user service that you want to enable
# and start automatically, use `default.target`
# For system level services, use `multi-user.target`
[Install]
WantedBy=multi-user.target
```

## Contributing

Fork this repo and checkout the *develop* branch.

```shell
$ git clone <your_forked_repo> -b develop
$ cd nginx-auth-server
```

Install the npm dependencies.

```shell
$ npm i
```

Build the JavaScript/TypeScript/SCSS stack once.

```shell
$ npm run build
```

Run the Go application

```shell
$ go build -o nginx-auth-server ./src/ && ./nginx-auth-server run
```

You can now point a nginx webserver to this auth-server. Refer to the nginx configuration in the *[Getting Started](#getting-started)* section.

If you want to make changes in the TypeScript/SCSS, you can run npm in *watch* mode:

```shell
$ npm run watch-ts
$ npm run watch-scss
```

**You have to restart the Go application after every change for the changes to take effect.**

## Documentation

The CLI and HTTP API documentation is available here: [https://burakkavak.github.io/nginx-auth-server/](https://burakkavak.github.io/nginx-auth-server/)

## Changelog

See [`CHANGELOG`](./CHANGELOG.md)

## Credits

- [etcd-io/bbolt](https://github.com/etcd-io/bbolt)
- [gin-gonic/gin](https://github.com/gin-gonic/gin)
- [go-ini/ini](https://github.com/go-ini/ini)
- [go-ldap/ldap](https://github.com/go-ldap/ldap)
- [pquerna/otp](https://github.com/pquerna/otp)
- [urfave/cli](https://github.com/urfave/cli)

## License

See [`LICENSE`](./LICENSE)
