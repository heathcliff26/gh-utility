###############################################################################
# BEGIN build-stage
# Compile the binary
FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.27.2 AS build-stage

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
FROM docker.io/library/alpine:3.24.2@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6 AS final-stage

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
