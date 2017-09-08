FROM golang:1.9
RUN apt-get update && apt-get install -y --no-install-recommends subversion && rm -rf /var/lib/apt/lists/*

COPY docker-entrypoint.sh /entrypoint.sh
WORKDIR /go/src/github.com/tomwei7/gosvn
ENV WORKDIR /go/src/github.com/tomwei7/gosvn
ENTRYPOINT ["/entrypoint.sh"]
