FROM golang:alpine AS builder
RUN apk add --no-cache nodejs npm
WORKDIR /build
COPY package.json package-lock.json ./
RUN npm install
COPY go.mod go.sum ./
RUN go get -d
COPY . .
RUN npm run build

FROM alpine
COPY --from=builder /build/caddy-analytics /usr/bin/caddy-analytics
CMD ["caddy-analytics"]
