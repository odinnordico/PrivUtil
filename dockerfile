FROM golang:tip-alpine3.23 as build
RUN apk update && apk add --no-cache make g++
RUN apk add nodejs npm
RUN mkdir PrivUtil
COPY . /PrivUtil/
WORKDIR /PrivUtil
RUN make build

ENTRYPOINT [ "./privutil" ]

FROM golang:tip-alpine3.23
COPY --from=build /PrivUtil/privutil /bin/privutil
ENTRYPOINT [ "privutil" ]
