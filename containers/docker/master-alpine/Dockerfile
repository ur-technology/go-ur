FROM alpine:3.4

RUN \
  apk add --update go git make gcc musl-dev         && \
  git clone https://github.com/ur-technology/go-ur && \
  (cd go-ur && make gur)                     && \
  cp go-ur/build/bin/gur /gur               && \
  apk del go git make gcc musl-dev                  && \
  rm -rf /go-ur && rm -rf /var/cache/apk/*

EXPOSE 9595
EXPOSE 19595

ENTRYPOINT ["/gur"]
