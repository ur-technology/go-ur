FROM alpine:3.3

RUN \
  apk add --update go git make gcc musl-dev         && \
  git clone https://github.com/urcapital/go-ur && \
  (cd go-ethereum && make geth)                     && \
  cp go-ethereum/build/bin/geth /geth               && \
  apk del go git make gcc musl-dev                  && \
  rm -rf /go-ethereum && rm -rf /var/cache/apk/*

EXPOSE 9595
EXPOSE 19595

ENTRYPOINT ["/geth"]
