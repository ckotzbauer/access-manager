FROM alpine:3.20@sha256:31687a2fdd021f85955bf2d0c2682e9c0949827560e1db546358ea094f740f12 as alpine

ARG TARGETARCH

RUN set -eux; \
    apk add -U --no-cache ca-certificates


FROM scratch

ARG TARGETOS
ARG TARGETARCH

COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY dist/access-manager_${TARGETOS}_${TARGETARCH}*/access-manager /usr/local/bin/access-manager

ENTRYPOINT ["/usr/local/bin/access-manager"]
