FROM golang:latest

WORKDIR /go/src/app
COPY . .

EXPOSE 53

RUN go get -d -v ./...
RUN go install -v

CMD ["app"]
