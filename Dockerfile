FROM alpine:3.17.0

ARG VERSION="0.0.8"

LABEL build_version="nginx-auth-server version ${VERSION}"
LABEL maintainer="burakkavak"

# ---------------- ENVIRONMENT VARIABLES START ----------------
ENV PATH="$PATH:/opt/nginx-auth-server"
ENV EXECUTABLE_NAME="nginx-auth-server"
ENV BASE_DIR="/opt/nginx-auth-server"

ENV SERVER_LISTEN_ADDRESS="0.0.0.0"
ENV SERVER_LISTEN_PORT=17397
ENV SERVER_DOMAIN="localhost"

ENV TLS_ENABLED="false"
ENV TLS_LISTEN_PORT=17760
ENV TLS_CERT_PATH="/opt/nginx-auth-server/certs/server.crt"
ENV TLS_CERT_KEY="/opt/nginx-auth-server/certs/server.key"

ENV COOKIES_LIFETIME=7
ENV COOKIES_SECURE="true"

ENV LDAP_ENABLED="false"
ENV LDAP_URL=""
ENV LDAP_ORGANIZATIONAL_UNIT="users"
ENV LDAP_DOMAIN_COMPONENTS=""

ENV RECAPTCHA_ENABLED="false"
ENV RECAPTCHA_SITE_KEY=""
ENV RECAPTCHA_SECRET_KEY=""
# ---------------- ENVIRONMENT VARIABLES END ----------------

RUN apkArch="$(apk --print-arch)"; \
    case "$apkArch" in \
      armhf) export ARCH='arm' ;; \
      x86) export ARCH='i386' ;; \
      aarch64) export ARCH='arm64' ;; \
      *) export ARCH='amd64' ;; \
    esac; \
    export EXECUTABLE_NAME="${EXECUTABLE_NAME}-linux-${ARCH}" && \
    apk add --no-cache --upgrade \
      wget \
      tar \
      libqrencode && \
    mkdir -p ${BASE_DIR} && \
    wget --content-disposition \
      -O ${BASE_DIR}/nginx-auth-server.tar.gz \
      https://github.com/burakkavak/nginx-auth-server/releases/download/${VERSION}/${EXECUTABLE_NAME}.tar.gz && \
    tar xf ${BASE_DIR}/nginx-auth-server.tar.gz -C ${BASE_DIR} && \
    rm ${BASE_DIR}/nginx-auth-server.tar.gz && \
    mv ${BASE_DIR}/${EXECUTABLE_NAME} ${BASE_DIR}/nginx-auth-server && \
    chown root:root ${BASE_DIR}/nginx-auth-server && \
    chmod +x ${BASE_DIR}/nginx-auth-server && \
    wget --content-disposition -O ${BASE_DIR}/docker_run.sh \
      https://raw.githubusercontent.com/burakkavak/nginx-auth-server/${VERSION}/scripts/docker_run.sh && \
    chmod +x ${BASE_DIR}/docker_run.sh


# ---------------- PORTS ----------------
EXPOSE ${SERVER_LISTEN_PORT}/tcp ${TLS_LISTEN_PORT}/tcp

ENTRYPOINT [ "docker_run.sh" ]
CMD [ "nginx-auth-server", "run" ]
