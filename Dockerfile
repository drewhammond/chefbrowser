ARG USERNAME=chefbrowser
ARG UID=1001
ARG GID=1001

FROM alpine:3.17 as alpine
ARG USERNAME
ARG UID
ARG GID
RUN apk add --update --no-cache ca-certificates shadow && \
    addgroup -g ${GID} ${USERNAME} && \
    adduser -u ${UID} -G ${USERNAME} --disabled-password --system ${USERNAME}

FROM scratch
ARG USERNAME
ARG UID
ARG GID
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=alpine /etc/passwd /etc/passwd
COPY chefbrowser /go/bin/chefbrowser
USER ${UID}:${GID}
EXPOSE 8080
ENTRYPOINT ["/go/bin/chefbrowser"]
