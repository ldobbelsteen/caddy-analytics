FROM node:alpine AS web-builder
WORKDIR /build
COPY web/package.json web/package-lock.json ./
RUN npm install
COPY web ./
RUN npm run build

FROM golang:alpine AS bin-builder
WORKDIR /build
COPY . .
RUN go get -d
RUN go build -o bin

FROM alpine
COPY --from=web-builder /build/dist /var/www/caddy-analytics
COPY --from=bin-builder /build/bin /usr/bin/caddy-analytics
CMD ["caddy-analytics", "--web", "/var/www/caddy-analytics"]
