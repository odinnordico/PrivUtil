FROM golang:tip-alpine3.23
RUN apk update && apk add --no-cache make g++
RUN mkdir privutil
RUN cd privutil
WORKDIR /privutil
COPY *.* .
RUN make build

ENTRYPOINT [ "/bin/sh", "privutil" ]