FROM node:15-alpine AS web
WORKDIR /build
COPY web/package.json web/package-lock.json ./
RUN npm install
COPY web ./
RUN npm run build

FROM golang:1.16-alpine AS bin
WORKDIR /build
COPY . .
RUN go get -d
COPY --from=web /build/web/dist ./web/dist
RUN go build -o bin

FROM alpine
COPY --from=bin /build/bin /usr/bin/caddy-analytics
CMD ["caddy-analytics"]
