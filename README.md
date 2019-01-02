HTTP over SSH
=============

[![CircleCI](https://circleci.com/gh/digineo/http-over-ssh.svg?style=svg)](https://circleci.com/gh/digineo/http-over-ssh)
[![Codecov](http://codecov.io/github/digineo/http-over-ssh/coverage.svg?branch=master)](http://codecov.io/github/digineo/http-over-ssh?branch=master)

This dynamic HTTP proxy tunnels your HTTP requests through SSH connections
using public key authentication. The intention to develop this program is
the requirement of polling [Prometheus exporters][promexp] through SSH.

[promexp]: https://prometheus.io/docs/instrumenting/exporters/

## Syntax

A proxy request looks like this:

    GET http://<jumphost>/<destination-host>/<destination-path> HTTP/1.1

You can override the SSH username by using HTTP Basic Auth.

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
    proxy_url: http://localhost:8080/
    metrics_path: /localhost:9100/metrics
    relabel_configs:
      - source_labels: ['__address__', '__metrics_path__']
        regex:        '(.+):\d+;/localhost:(\d+)/.*'
        replacement:  '$1:$2'
        target_label: 'instance'
    static_configs:
      - targets:
        - www.example.com:22
        - mail.example.com:22
```

### Authorized Keys (OpenSSH)

To restrict an SSH key to only forward connections to `localhost:9100`, append to the `~/.ssh/authorized_keys`:

```
restrict,port-forwarding,permitopen="localhost:9100" ssh-ed25519 <the-key> prometheus@example.com
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

- [ ] verify SSH host keys
- [ ] add support for SSH certificates
- [ ] clean up idle ssh connections
- [ ] support for unix sockets

## License

MIT Licence. Copyright 2018, Digineo GmbH
