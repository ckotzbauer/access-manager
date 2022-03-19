FROM alpine:3.15@sha256:d6d0a0eb4d40ef96f2310ead734848b9c819bb97c9d846385c4aca1767186cd4 as alpine

ARG TARGETARCH

RUN set -eux; \
    apk add -U --no-cache ca-certificates


FROM scratch

ARG TARGETOS
ARG TARGETARCH

COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY dist/access-manager_${TARGETOS}_${TARGETARCH}/access-manager /usr/local/bin/access-manager

ENTRYPOINT ["/usr/local/bin/access-manager"]
