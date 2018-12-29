HTTP over SSH
=============

This dynamic HTTP proxy tunnels your HTTP requests through SSH connections
using public key authentication. The intention to develop this program is
the requirement of polling [Prometheus exporters][promexp] through SSH.

> **It is not recommended for production use, yet.**

[promexp]: https://prometheus.io/docs/instrumenting/exporters/

## Syntax

    GET http://<jumphost>/<destination-host>/<destination-path>


## Usage

After installation (see below), start the proxy on `localhost:8000`:

```console
$ http-over-ssh -listen 127.0.0.1:8000
```

For a full list of options run `http-over-ssh -help`.

### Prometheus Scraper

Assuming this proxy runs on the same machine as Prometheus on `localhost:8080`
and you want to scrape to remote hosts running prometheus exporters on `localhost:9100`,
simply add to your scrape configs:

```yaml
  - job_name: 'node-exporter'
    static_configs:
      - targets:
        - www.example.com:22
        - mail.example.com:22
    proxy_url: http://localhost:8080/
    metrics_path: /localhost:9100/metrics
```


## Installation

If you have the Go toolchain installed, a simple

```console
$ go get github.com/digineo/http-over-ssh
```

will place a `http-over-ssh` binary in `$GOPATH/bin/`.

Alternatively, you may download a pre-built binary from the Github
[release page][releases] and extract the binary into your `$PATH`.

[releases]: https://github.com/digineo/http-over-ssh/releases

## Next steps

- [ ] stability improvements
- [ ] clean up idle ssh connections
- [ ] support for unix sockets

## License

MIT Licence. Copyright 2018, Digineo GmbH
