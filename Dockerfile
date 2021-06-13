FROM golang:1.16-alpine AS builder
RUN apk add --no-cache nodejs npm make
WORKDIR /build
COPY . .
RUN make

FROM alpine
COPY --from=builder /build/bin/caddy-analytics /usr/bin/caddy-analytics
CMD ["caddy-analytics"]
