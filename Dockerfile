FROM alpine:3.19@sha256:c5b1261d6d3e43071626931fc004f70149baeba2c8ec672bd4f27761f8e1ad6b as alpine

ARG TARGETARCH

RUN set -eux; \
    apk add -U --no-cache ca-certificates


FROM scratch

ARG TARGETOS
ARG TARGETARCH

COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY dist/access-manager_${TARGETOS}_${TARGETARCH}*/access-manager /usr/local/bin/access-manager

ENTRYPOINT ["/usr/local/bin/access-manager"]
