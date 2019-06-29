FROM golang:latest

WORKDIR /go/src/app
COPY . .

EXPOSE 53
ENV FASTDNS_CONFIG=/app/config.json
ENV FASTDNS_DNS1=1.1.1.1
ENV FASTDNS_DNS2=1.0.0.1

RUN go get -d -v ./...
RUN go install -v

CMD ["app"]
