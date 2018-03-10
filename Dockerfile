FROM golang:latest

ADD . /go/src/app
WORKDIR /go/src
RUN go get app
RUN go install app
ENTRYPOINT /go/bin/app
