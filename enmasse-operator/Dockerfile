FROM alpine:3.8

RUN apk upgrade --update --no-cache

ARG version
ARG commit
ENV VERSION=${version} COMMIT=${commit}

USER nobody

ADD enmasse-controller-manager /usr/local/bin/enmasse-controller-manager
