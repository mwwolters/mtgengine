FROM golang:alpine

RUN apk update \
  && apk add git

COPY ./collection/ /go/src/mtgengine/collection/
COPY ./engine/ /go/src/mtgengine/engine/
COPY ./srv/ /go/src/mtgengine/srv/
COPY ./proto/ /go/src/mtgengine/proto/
COPY ./server.go /go/src/mtgengine/

# Don't do this in production! Use vendoring instead.
RUN go get -v mtgengine

ENV GOBIN /go/bin

RUN go install /go/src/mtgengine/server.go


ENTRYPOINT ["/go/bin/server", "--port", "50051", "--sets", "/tmp/sets/AllSets.json"]
