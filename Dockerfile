#Build Stage
FROM golang:1.23.1

ENV GIN_MODE release

WORKDIR /go/src/app

# install needed library or anthing else
RUN go install github.com/air-verse/air@latest

#Final Stage
COPY ./app .
CMD air