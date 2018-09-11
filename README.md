HTTP over SSH
=============

This dynamic HTTP proxy tunnels your HTTP requests through SSH connections using public key authentication.
It is not recommended for production use.

The intention to develope this program is the requirement of polling [Prometheus exporters](https://prometheus.io/docs/instrumenting/exporters/) through SSH.

## Syntax

    http://<proxy-address>/<ssh-address>/<destination-host>/<destination-path>


## Example

Your proxy is reachable at `localhost:8000` and you want to fetch http://example.com/index.html via `jumphost.tld:22`.
Then just fetch:

    http://localhost:8080/jumphost.tld:22/example.com:80/index.html


## Next steps

- [ ] stability improvements
- [ ] clean up idle ssh connections
- [ ] support for unix sockets
