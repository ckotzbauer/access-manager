FROM alpine:3.12

ENV USER_UID=1001 \
    USER_NAME=access-manager

# install operator binary
COPY bin/manager /usr/local/bin/access-manager

RUN echo "${USER_NAME}:x:${USER_UID}:0:${USER_NAME} user:${HOME}:/sbin/nologin" >> /etc/passwd && \
    mkdir -p "${HOME}" && \
    chown "${USER_UID}:0" "${HOME}" && \
    chmod ug+rwx "${HOME}"

ENTRYPOINT ["/usr/local/bin/access-manager"]

USER ${USER_UID}
