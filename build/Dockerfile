FROM golang:1.24-alpine AS build

RUN apk add --no-cache git

ENV BIN_FILE /opt/anti-brute-force
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o ${BIN_FILE} ./cmd/

FROM alpine:3.19

RUN apk add --no-cache tzdata ca-certificates

EXPOSE 8080
COPY --from=build /opt/anti-brute-force /opt/anti-brute-force
ENTRYPOINT ["/opt/anti-brute-force"]