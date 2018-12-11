HTTP over SSH
=============

This dynamic HTTP proxy tunnels your HTTP requests through SSH connections
using public key authentication. The intention to develop this program is
the requirement of polling [Prometheus exporters][promexp] through SSH.

> **It is not recommended for production use, yet.**

[promexp]: https://prometheus.io/docs/instrumenting/exporters/

## Syntax

    http://<proxy-address>/<ssh-address>/<destination-host>/<destination-path>


## Usage

After installation (see below), start the proxy on `localhost:8000`:

```console
$ http-over-ssh -listen 127.0.0.1:8000
```

For a full list of options run `http-over-ssh -help`.

If you want to fetch http://example.com/index.html via `root@jumphost.tld:22`,
just fetch this URL instead:

    http://localhost:8080/root@jumphost.tld:22/example.com:80/index.html

For `<ssh-address>`, the defaults for username and port are "root" and 22.
The following fetch URL is hence equivalent:

    http://localhost:8080/jumphost.tld/example.com:80/index.html

For the `<destination-host>` is currently only HTTP allowed. This might
change in the future, but requires a change in the fetch URL syntax.

Parsing IPv6 addresses for both `<ssh-address>` and `<destination-host>`
is currently buggy as well.

Please [open an issue][issues] if you need one those features.

[issues]: https://github.com/digineo/http-over-ssh/issues


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
