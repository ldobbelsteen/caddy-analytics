build:
	npm install --prefix web
	npm run build --prefix web
	go get -d
	go build -o bin/caddy-analytics
