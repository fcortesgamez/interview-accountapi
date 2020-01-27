FROM golang:1.13-alpine

# Install git and rest of build dependencies
RUN set -ex; \
    apk update; \
    apk add --no-cache git bash curl make

RUN apk --no-cache add ca-certificates wget bash \
    && wget -q -O /etc/apk/keys/sgerrand.rsa.pub https://alpine-pkgs.sgerrand.com/sgerrand.rsa.pub \
    && wget https://github.com/sgerrand/alpine-pkg-glibc/releases/download/2.29-r0/glibc-2.29-r0.apk \
    && apk add glibc-2.29-r0.apk

# Set working directory
WORKDIR /go/src/github.com/fcortesgamez/interview-accountapi/
