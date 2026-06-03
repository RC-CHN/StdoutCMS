FROM golang:1.26-alpine AS builder
WORKDIR /build
ARG GOPROXY
ENV GOPROXY=${GOPROXY}
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -ldflags "-X stdoutcms/internal/handler.Version=${VERSION}" -o server ./cmd/server

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /build/server .
COPY --from=builder /build/migrations ./migrations
EXPOSE 8080
ENTRYPOINT ["./server"]
