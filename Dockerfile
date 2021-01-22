FROM golang:1.15.7-buster as builder

ARG version
ENV VERSION=$version

WORKDIR /go/src/app
COPY . .
RUN make manager


FROM alpine:3.13

ENV USER_UID=1001 \
    USER_NAME=access-manager

RUN echo "${USER_NAME}:x:${USER_UID}:0:${USER_NAME} user:${HOME}:/sbin/nologin" >> /etc/passwd && \
    mkdir -p "${HOME}" && \
    chown "${USER_UID}:0" "${HOME}" && \
    chmod ug+rwx "${HOME}"

COPY --from=builder /go/src/app/bin/manager /usr/local/bin/access-manager

ENTRYPOINT ["/usr/local/bin/access-manager"]
USER ${USER_UID}
