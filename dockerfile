FROM golang:tip-alpine3.23
RUN apk update && apk add --no-cache make g++
RUN apk add nodejs npm
RUN mkdir PrivUtil
COPY . /PrivUtil/
WORKDIR /PrivUtil
RUN make build

ENTRYPOINT [ "./privutil" ]