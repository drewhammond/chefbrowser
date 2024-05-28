ARG BASE_IMAGE=alpine:3.18
ARG USERNAME=chefbrowser
ARG UID=1001
ARG GID=1001

FROM --platform=$BUILDPLATFORM $BASE_IMAGE as cb-base
ARG USERNAME
ARG UID
ARG GID
RUN apk add --update --no-cache ca-certificates shadow && \
    addgroup -g ${GID} ${USERNAME} && \
    adduser -u ${UID} -G ${USERNAME} --disabled-password --system ${USERNAME}

###################
# UI build stage
###################
FROM --platform=$BUILDPLATFORM node:20-alpine3.18 AS ui-builder
WORKDIR /src
COPY ["ui/package.json", "ui/yarn.lock", "./"]
RUN yarn install --network-timeout 200000 && \
    yarn cache clean

COPY ["ui/", "."]

ARG TARGETARCH
RUN HOST_ARCH=$TARGETARCH NODE_ENV='production' NODE_ONLINE_ENV='online' NODE_OPTIONS=--max_old_space_size=8192 yarn run build

###################
# Go build stage
###################
FROM --platform=$BUILDPLATFORM golang:1.22.2 as go-builder
WORKDIR /go/src/github.com/drewhammond/chefbrowser
COPY go.* ./
RUN go mod download

COPY . .
COPY --from=ui-builder /src/dist /go/src/github.com/drewhammond/chefbrowser/ui/dist
ARG TARGETOS
ARG TARGETARCH
ARG RELEASE=dev
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH RELEASE=$RELEASE make build-backend

###################
# Final stage
###################
FROM scratch
ARG UID
ARG GID
COPY --from=cb-base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=cb-base /etc/passwd /etc/passwd
COPY --from=go-builder /go/src/github.com/drewhammond/chefbrowser/dist/chefbrowser /usr/local/bin/
USER ${UID}:${GID}
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/chefbrowser"]
