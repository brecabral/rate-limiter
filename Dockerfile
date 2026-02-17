FROM golang:1.25 AS build
WORKDIR /wd
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app ./cmd/server

FROM alpine:3.23
WORKDIR /wd
COPY --from=build /wd/app .
EXPOSE 8080
ENTRYPOINT ["./app"]
