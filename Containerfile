# syntax=docker/dockerfile:1
FROM docker.io/library/golang:1.25-alpine AS builder
RUN apk add --no-cache curl make
WORKDIR /src
ENV CGO_ENABLED=0
ENV GOTOOLCHAIN=auto
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN make swaggerUI
RUN go build -o /knxctl ./cmd/knxctl

FROM docker.io/library/alpine:3.23
RUN apk add --no-cache ca-certificates
COPY --from=builder /knxctl /usr/local/bin/knxctl
ENTRYPOINT ["knxctl"]
