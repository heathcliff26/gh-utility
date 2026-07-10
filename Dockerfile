###############################################################################
# BEGIN build-stage
# Compile the binary
FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.26.5 AS build-stage

ARG BUILDPLATFORM
ARG TARGETARCH

WORKDIR /app

COPY vendor ./vendor

COPY go.mod go.sum ./

COPY hack ./hack

COPY cmd ./cmd

COPY pkg ./pkg

RUN GOOS=linux GOARCH="${TARGETARCH}" hack/build.sh

#
# END build-stage
###############################################################################

###############################################################################
# BEGIN final-stage
# Create final docker image
FROM docker.io/library/alpine:3.24.1@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b AS final-stage

LABEL   org.opencontainers.image.authors="heathcliff@heathcliff.eu" \
        org.opencontainers.image.description="CLI tool to interact with the GitHub API as an app" \
        org.opencontainers.image.source="https://github.com/heathcliff26/gh-utility" \
        org.opencontainers.image.licenses="Apache-2.0" \
        org.opencontainers.image.title="gh-utility"

RUN apk add --no-cache github-cli jq && adduser -D github

COPY --from=build-stage /app/bin/gh-utility /usr/local/bin/gh-utility

WORKDIR /home/github

USER github:github

#
# END final-stage
###############################################################################
