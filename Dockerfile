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
RUN CGO_ENABLED=0 go build -o bin

FROM scratch
COPY --from=web-builder /build/dist /web
COPY --from=bin-builder /build/bin /caddy-analytics
CMD ["/caddy-analytics", "--web", "/web"]
