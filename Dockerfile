# To build:
# $ docker run --rm -v $(pwd):/go/src/github.com/micahhausler/container-tx -w /go/src/github.com/micahhausler/container-tx golang:1.7  go build -v -a -tags netgo -installsuffix netgo -ldflags '-w'
# $ docker build -t micahhausler/container-tx .
#
# To run:
# $ docker run micahhausler/container-tx

FROM busybox

MAINTAINER Micah Hausler, <hausler.m@gmail.com>

COPY container-tx /bin/container-tx
RUN chmod 755 /bin/container-tx

ENTRYPOINT ["/bin/container-tx"]
