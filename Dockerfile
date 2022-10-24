FROM golang:1.19 AS builder

RUN go version
ENV GOPATH=/

COPY ./ ./

RUN go mod download
RUN go build -o WeatherServiceAPI ./cmd/main/app.go

EXPOSE 8090

CMD ["./WeatherServiceAPI"]