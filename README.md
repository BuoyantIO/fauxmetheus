# fauxmetheus
Serves fake Prometheus metrics which simulate a Linkerd cluster

## Usage

```console
> go build -o fauxmetheus *.go
> ./fauxmetheus tiny.json
Serving /metrics on :4191
```

## Details

Fauxmetheus reads its configuration from a json file (see `tiny.json` and
`medium.json` for examples) and then serves a `/metrics` endpoint on port
`4191`.  This endpoint returns Prometheus formatted metrics, simulating the
metrics that would be gathered from a cluster of Linkerd-proxy processes.
The following metrics are included:

* response_total
* response_latency_ms_bucket
* tcp_open_connections
* tcp_read_bytes_total
* tcp_write_bytes_total
