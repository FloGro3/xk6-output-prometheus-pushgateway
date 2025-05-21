package pushgateway

import (
	"fmt"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.k6.io/k6/metrics"
)

// CollectorResolverSummary is an interface to resolve the various types of the [metrics.Metric]
// to the [prometheus.Collector].
// Respective [k6 metric type] are solved by following the [conversion rule] which
// the [xk6-output-prometheus-remote] extension applies.
//
// [k6 metric type]: https://k6.io/docs/using-k6/metrics/#metric-types
// [conversion rule]: https://k6.io/blog/k6-loves-prometheus/#mapping-k6-metrics-types
// [xk6-output-prometheus-remote]: https://github.com/grafana/xk6-output-prometheus-remote
type CollectorResolverSummary func(sample metrics.Sample, prefix string) []prometheus.Collector

// CollectorResolverSummary is a factory method to create the [ColloectorResolver] implementation
// corresponding to the given [k6 metric type].
//
// Example use case:
//
//	collectorResolver := collector_resolver.CollectorResolverSummary(sample.Metric.Type)
//	collectors := collectorResolverSummary(sample.Metric, time.Now())
//
// [k6 metric type]: https://k6.io/docs/using-k6/metrics/#metric-types
func CreateResolverSummary(t metrics.MetricType) CollectorResolverSummary {
	var resolver CollectorResolverSummary
	switch t {
	case metrics.Counter:
		resolver = resolveCounterSummary
	case metrics.Gauge:
		resolver = resolveGaugeSummary
	case metrics.Rate:
		resolver = resolveRateSummary
	case metrics.Trend:
		resolver = resolveTrendSummary
	}
	return resolver
}

func resolveCounterSummary(sample metrics.Sample, prefix string) []prometheus.Collector {
	counter := prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Namespace: 	prefix,
			Name:      	sample.Metric.Name,
		},
		func() float64 {
			counterSink := sample.Metric.Sink.(*metrics.CounterSink)
			return sample.Metric.Sink.Format(sample.GetTime().Sub(counterSink.First))["rate"]
		},
	)
	return []prometheus.Collector{counter}
}

func resolveGaugeSummary(sample metrics.Sample, prefix string) []prometheus.Collector {
	gauge := prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: 	prefix,
			Name:      	sample.Metric.Name,
		},
		func() float64 {
			return sample.Value
		},
	)
	return []prometheus.Collector{gauge}
}

func resolveRateSummary(sample metrics.Sample, prefix string) []prometheus.Collector {
	gauge := prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: 	prefix,
			Name:      	sample.Metric.Name,
		},
		func() float64 {
			return sample.Value
		},
	)
	return []prometheus.Collector{gauge}
}

func resolveTrendSummary(sample metrics.Sample, prefix string) []prometheus.Collector {
	sink := sample.Metric.Sink.Format(time.Duration(0))

	collectors := make([]prometheus.Collector, 0)
	for k, v := range sink {
		// Remove "(" and ")" from the name of prometheus collector
		// Becuase these are not acceptable as collector name.
		suffix := strings.ReplaceAll(strings.ReplaceAll(k, "(", ""), ")", "")

		name := fmt.Sprintf("%s_%s", sample.Metric.Name, suffix)
		gauge := prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: 	prefix,
				Name:      	name,
			},
		)
		gauge.Set(v)
		collectors = append(collectors, gauge)
	}
	return collectors
}
