FROM node AS web-builder
WORKDIR /build
COPY web/package.json web/package-lock.json ./
RUN npm install
COPY web ./
RUN npm run build

FROM golang AS bin-builder
WORKDIR /build
COPY . .
RUN go get -d
RUN go build -o bin

FROM scratch
COPY --from=web-builder /build/dist /app/web
COPY --from=bin-builder /build/bin /app/caddy-analytics
CMD ["/app/caddy-analytics", "--web", "/app/web", "--geo", "/app/geo.mmdb"]
