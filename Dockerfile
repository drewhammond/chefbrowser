FROM alpine:3.16 as alpine
RUN apk add --update --no-cache ca-certificates

FROM scratch
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=alpine /etc/passwd /etc/passwd
COPY chefbrowser /go/bin/chefbrowser
USER nobody
ENV GIN_MODE=release
EXPOSE 8080
ENTRYPOINT ["/go/bin/chefbrowser"]
