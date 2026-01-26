FROM golang:tip-alpine3.23
RUN apk update && apk add --no-cache make g++
RUN apk add nodejs npm
RUN mkdir PrivUtil
RUN ls -lahrt
COPY . /PrivUtil/
WORKDIR /PrivUtil
RUN pwd 
RUN ls -lahr
RUN make build

ENTRYPOINT [ "/bin/sh", "privutil" ]