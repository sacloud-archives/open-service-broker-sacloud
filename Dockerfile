FROM golang:1.10 AS builder
MAINTAINER Kazumichi Yamamoto <yamamoto.febc@gmail.com>
LABEL MAINTAINER 'Kazumichi Yamamoto <yamamoto.febc@gmail.com>'

RUN  apt-get update && apt-get -y install \
        bash \
        git  \
        make \
        zip  \
      && apt-get clean \
      && rm -rf /var/cache/apt/archives/* /var/lib/apt/lists/*

ADD . /go/src/github.com/sacloud/open-service-broker-sacloud
WORKDIR /go/src/github.com/sacloud/open-service-broker-sacloud
RUN ["make", "clean", "build"]

#----------

FROM alpine:3.7
MAINTAINER Kazumichi Yamamoto <yamamoto.febc@gmail.com>
LABEL MAINTAINER 'Kazumichi Yamamoto <yamamoto.febc@gmail.com>'

RUN set -x && apk add --no-cache --update ca-certificates
COPY --from=builder /go/src/github.com/sacloud/open-service-broker-sacloud/bin/open-service-broker-sacloud /usr/local/bin/
RUN chmod +x /usr/local/bin/open-service-broker-sacloud
ENTRYPOINT ["/usr/local/bin/open-service-broker-sacloud"]
