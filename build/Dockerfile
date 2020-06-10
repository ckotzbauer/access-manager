FROM alpine:3.12

ENV OPERATOR=/usr/local/bin/access-manager \
    USER_UID=1001 \
    USER_NAME=access-manager

# install operator binary
COPY build/_output/bin/access-manager ${OPERATOR}

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
