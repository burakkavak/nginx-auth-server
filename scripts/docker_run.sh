#!/bin/sh

# This script is used inside the docker container to bootstrap the application.

# Rewrite/create config with current environment.
{
  echo "[Server]"
  echo "listen_address = $SERVER_LISTEN_ADDRESS"
  echo "listen_port = $SERVER_LISTEN_PORT"
  echo "domain = $SERVER_DOMAIN"

  echo "[TLS]"
  echo "enabled = $TLS_ENABLED"
  echo "listen_port = $TLS_LISTEN_PORT"
  echo "cert_path = $TLS_CERT_PATH"
  echo "key_path = $TLS_CERT_KEY"

  echo "[Cookies]"
  echo "lifetime = $COOKIES_LIFETIME"
  echo "secure = $COOKIES_SECURE"

  echo "[LDAP]"
  echo "enabled = $LDAP_ENABLED"
  echo "url = $LDAP_URL"
  echo "organizational_unit = $LDAP_ORGANIZATIONAL_UNIT"
  echo "domain_components = $LDAP_DOMAIN_COMPONENTS"

  echo "[Recaptcha]"
  echo "enabled = $RECAPTCHA_ENABLED"
  echo "site_key = $RECAPTCHA_SITE_KEY"
  echo "secret_key = $RECAPTCHA_SECRET_KEY"
} > "${BASE_DIR}/config.ini"

# execute Dockerfile 'CMD'
exec "$@"
