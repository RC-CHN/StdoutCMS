FROM golang:1.26-alpine AS builder
WORKDIR /build
ARG GOPROXY
ENV GOPROXY=${GOPROXY}
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
COPY VERSION /build/VERSION
RUN CGO_ENABLED=0 go build -ldflags "-X stdoutcms/internal/handler.Version=$(cat VERSION)" -o server ./cmd/server

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /build/server .
COPY --from=builder /build/migrations ./migrations
EXPOSE 8080
ENTRYPOINT ["./server"]
