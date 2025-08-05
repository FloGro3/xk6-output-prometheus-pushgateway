xk6-output-prometheus-pushgateway
===
This is a k6 extension for publishing test-run metrics to Prometheus via [Pushgateway](https://prometheus.io/docs/instrumenting/pushing/).\
This extension is fully inspired by [xk6-output-prometheus-remote](https://github.com/grafana/xk6-output-prometheus-remote).\
There might be a circumstance not to enable the "[Remote Write](https://prometheus.io/docs/practices/remote_write/)" feature on your Prometheus instance. In that case, the [Pushgateway](https://prometheus.io/docs/instrumenting/pushing/) and this extension are possibly be an alternative solution.
The extension was optimized to work with distributed k6 tests using [k6 operator](https://github.com/grafana/k6-operator). The metrics can be aggregated later on in grafana for the same k6 jobs. 

## Development commands
- `go mod tidy` - Run this if versions have changed in the `go.mod` file to clean up unused dependencies.
- `go test -cover ./...` - Executes tests and shows test coverage for all packages.

## Usage
```sh
% xk6 build --with github.com/FloGro3/xk6-output-prometheus-pushgateway@latest
% K6_PUSHGATEWAY_URL=http://localhost:9091 \
K6_JOB_NAME=k6_load_testing \
./k6 run \
./script.js \
-o output-prometheus-pushgateway
```

# Label-Segregation

By default the reported metrics are splitted up by the scenario name which is exported as a tag by k6. This means the scenario name is injected into the metrics name. This makes it possible to filter metrics by scenarios. 

Via

```
K6_LABEL_SEGREGATION="name, method"
```

additional splitting of metrics can be configured. For possible values check [K6 tags and groups](https://grafana.com/docs/k6/latest/using-k6/tags-and-groups/).

# Metrics prefix

It is possible to configure this output to expose the time series with a prefix.

Configure it with the following environment variable:

```
K6_PUSHGATEWAY_NAMESPACE=k6 k6 run ...
```

# Configuration options

| Key | Default | Description |
|-----|---------|-------------|
| K6_PUSH_INTERVAL | `5` | Intervall in seconds for sending metrics to pushgateway |
| K6_PUSHGATEWAY_URL | `"http://localhost:9091"` | URL of pushgateway to whitch the metrics are send |
| K6_PUSHGATEWAY_NAMESPACE |  | Setting namespace of the metric |
| K6_JOB_NAME | `"k6_load_testing"` | Sets a job tag as a grouping key for the metrcis send to pusgateway |
| K6_LABEL_SEGREGATION | `"scenario"` | Defines labes by which the metrics are split up |