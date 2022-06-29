FROM golang:alpine


COPY src /go/src/logsapi/src
WORKDIR /go/src/logsapi/src/logsapi


RUN go get /go/src/logsapi/src/logsapi
RUN go build -o ./logsapi /go/src/logsapi/src/logsapi

ENTRYPOINT ["./logsapi"]
 