FROM alpine:3.4

ADD . /go-ur
RUN \
  apk add --update git go make gcc musl-dev         && \
  (cd go-ur && make gur)                     && \
  cp go-ur/build/bin/gur /gur               && \
  apk del git go make gcc musl-dev                  && \
  rm -rf /go-ur && rm -rf /var/cache/apk/*

EXPOSE 9595
#EXPOSE 30303
EXPOSE 19595

ENTRYPOINT ["/gur"]
