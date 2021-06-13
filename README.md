# Caddy Analytics
Caddy web server log analyzer with web interface written in Go. It is akin to [GoAccess](https://github.com/allinurl/goaccess), but specialized for Caddy.

## Building
A Makefile is included to build a binary. It first bundles the web interface and then builds a Go binary with the web interface embedded. To build, simply run:

```
make
```

The executable binary can then be found in the `bin` directory. There is also a Dockerfile included with which a Docker image can be built by running:

```
docker build --tag caddy-analytics https://github.com/ldobbelsteen/caddy-analytics.git
```

Pre-built images can be found [here](https://hub.docker.com/r/ldobbelsteen/caddy-analytics).

## Usage
```
caddy-analytics
	[--geo <license-key>]
	[--logs <log-directory>]
	[--port <listen-port>]
	[--cache <seconds>]
```

`--geo` specifies the MaxMind license key used for downloading and maintaining a GeoIP country database. Creating a license key is free, but requires an account. This option is mandatory.

`--logs` specifies the directory to which Caddy's logs are written. This directory can contain both active logs and rolled/compressed logs. Only logs with extensions `.log` and `.log.gz` are analyzed. Defaults to `/var/log/caddy`.

`--port` specifies the port on which the web interface be served. Defaults to `5734`.

`--cache` specifies the number of seconds to cache a log analysis before discarding. Defaults to `10`.

These options can also be passed through environment variables with their counterparts `GEO`, `LOGS`, `PORT` and `CACHE`. Environment variables have priority over command line arguments. The Docker image uses these environment variables for configuration. The simplest usage of the Docker image is:

```
docker run \
	--publish 5734:5734 \
	--volume /path/to/logs:/var/log/caddy \
	--env GEO=your-license-key \
	caddy-analytics
```

## Logging
Caddy doesn't log to disk by default. To enable logging, take a look at the docs for your config format too see how ([Caddyfile](https://caddyserver.com/docs/caddyfile/directives/log), [JSON](https://caddyserver.com/docs/json/logging/)). This is a simple example to enable logging for specific hosts in the Caddyfile using snippets:

```
(logging) {
	log {
		output file /var/log/caddy/access.log
	}
}

example.com {
	import logging
	reverse_proxy localhost:8080
}
```
