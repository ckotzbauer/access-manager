FROM gcr.io/distroless/base

ARG TARGETOS
ARG TARGETARCH

COPY dist/access-manager_${TARGETOS}_${TARGETARCH}/access-manager /usr/local/bin/access-manager

ENTRYPOINT ["/usr/local/bin/access-manager"]
