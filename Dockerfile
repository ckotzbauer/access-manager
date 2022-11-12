FROM alpine:3.16@sha256:b95359c2505145f16c6aa384f9cc74eeff78eb36d308ca4fd902eeeb0a0b161b as alpine

ARG TARGETARCH

RUN set -eux; \
    apk add -U --no-cache ca-certificates


FROM scratch

ARG TARGETOS
ARG TARGETARCH

COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY dist/access-manager_${TARGETOS}_${TARGETARCH}*/access-manager /usr/local/bin/access-manager

ENTRYPOINT ["/usr/local/bin/access-manager"]
