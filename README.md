# nginx-auth-server

A lightweight authentication server designed to be used in conjunction with nginx 'http_auth_request_module'. nginx-auth-server provides an additional authentication layer that is useful for reverse proxy scenarios, where the proxy does not support user authentication.

## Table of Contents

- [Demo](#demo)
- [Features](#features)
- [Getting Started](#getting-started)
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

Download the appropriate binary from the [Releases](https://github.com/burakkavak/nginx-auth-server/releases) section.

Download the current [config.ini](https://raw.githubusercontent.com/burakkavak/nginx-auth-server/master/config.ini) into the same directory:

```shell
$ wget --content-disposition https://raw.githubusercontent.com/burakkavak/nginx-auth-server/master/config.ini
```

Run the server:
```shell
$ ./nginx-server-auth run
```

For user management (adding/removing users) refer to the CLI usage information:
```shell
$ ./nginx-server-auth help
$ ./nginx-server-auth user add --username foo --password bar --otp
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

  # these are handled by nginx-server-auth as part of the auth routines
  location ~ ^/(login|logout|whoami)$ {
    proxy_pass http://localhost:17397;

    proxy_set_header X-Original-URI $request_uri;
    proxy_set_header X-Original-Remote-Addr $remote_addr;
    proxy_set_header X-Original-Host $host;
  }

  # static nginx-server-auth assets (css, js, ...)
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
WorkingDirectory=/var/www/nginx-auth-server/build
ExecStart=/var/www/nginx-auth-server/build/nginx-auth-server run
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