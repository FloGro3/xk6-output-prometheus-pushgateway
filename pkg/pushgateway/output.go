package pushgateway

import (
	"fmt"
	"sync"
	"time"
	"strings"

	collector_resolver "github.com/FloGro3/xk6-output-prometheus-pushgateway/pkg/pushgateway/collector_resolver"

	"github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"go.k6.io/k6/metrics"
	"go.k6.io/k6/output"
)

// Output implements the lib.Output interface
type Output struct {
	output.SampleBuffer

	config          Config
	periodicFlusher *output.PeriodicFlusher
	logger          logrus.FieldLogger
	waitGroup       sync.WaitGroup
}

var _ output.Output = new(Output)

// New creates an instance of the collector
func New(p output.Params) (*Output, error) {
	conf, err := NewConfig(p)
	if err != nil {
		return nil, err
	}
	// Some setupping code

	return &Output{
		config: conf,
		logger: p.Logger,
	}, nil
}

func (o *Output) Description() string {
	return fmt.Sprintf("pushgateway: %s, job: %s, labels: %s", o.config.PushGWUrl, o.config.JobName, o.config.LabelSegregation)
}

func (o *Output) Stop() error {
	o.logger.Debug("Stopping...")
	defer o.logger.Debug("Stopped!")
	o.periodicFlusher.Stop()
	o.waitGroup.Wait()

	return nil
}

func (o *Output) Start() error {
	o.logger.Debug("Starting...")

	// Here we should connect to a service, open a file or w/e else we decided we need to do

	pf, err := output.NewPeriodicFlusher(o.config.PushInterval, o.flushMetrics)
	if err != nil {
		return err
	}
	o.logger.Debug("Started!")
	o.periodicFlusher = pf

	return nil
}

func (o *Output) flushMetrics() {
	samplesContainers := o.GetBufferedSamples()

	o.waitGroup.Add(1)
	go func() {
		defer o.waitGroup.Done()

		//sampleMap := extractPushSamples(sampleContainers)
		for _, samplesContainer := range samplesContainers {
			samples := samplesContainer.GetSamples()

			o.logger.WithFields(dumpk6Sample(samples)).Debug("Dump k6 samples.")
			collectors := convertk6SamplesToPromCollectors(o, samples, o.config.Namespace)

			registry := prometheus.NewPedanticRegistry()
			registry.MustRegister(collectors...)
			o.logger.WithFields(dumpPrometheusCollector(registry)).Debug("Dump collectors.")

			pusher := push.New(o.config.PushGWUrl, o.config.JobName)
			instance := ""
			//get job_name tag from sample (should be same accross all samples) as instance grouping value
			for _, value := range samples {
				tagSet := value.GetTags()
				if (tagSet != nil) {
					if tag, exists := tagSet.Get("job_name"); exists {
						instance = tag
						break
					}
				}
			}
			if err := pusher.Gatherer(registry).Grouping("instance", instance).Add(); err != nil {
				o.logger.
					Errorf("Could not add to Pushgateway: %v", err)
			}
		}
	}()
}

func convertk6SamplesToPromCollectors(o *Output, samples []metrics.Sample, prefix string) []prometheus.Collector {
	collectors := make([]prometheus.Collector, 0)
	for _, sample := range samples {
		resolver := collector_resolver.CreateResolver(sample.Metric.Type)
		resolverSummary := collector_resolver.CreateResolverSummary(sample.Metric.Type)
		subsystem := getSubsystem(o, sample)
		collectors = append(collectors, resolver(sample, subsystem, prefix)...)
		collectors = append(collectors, resolverSummary(sample, prefix)...)
	}
	return collectors
}

func dumpk6Sample(samples []metrics.Sample) logrus.Fields {
	var value float64
	t := time.Since(time.Now())
	fields := logrus.Fields{}
	for _, sample := range samples {
		switch sample.Metric.Type {
		case metrics.Counter:
			value = sample.Metric.Sink.Format(t)["count"]
		case metrics.Gauge:
			value = sample.Metric.Sink.Format(t)["value"]
		case metrics.Rate:
			value = sample.Metric.Sink.Format(t)["rate"]
		}
		fields[sample.Metric.Name] = map[string]interface{}{
			"sample_value": sample.Value,
			"sink_value":   value,
			"name":         sample.Metric.Name,
			"type":         sample.Metric.Type,
		}
	}
	return fields
}

func dumpPrometheusCollector(reg *prometheus.Registry) logrus.Fields {
	metricFamilies, _ := prometheus.Gatherers{reg}.Gather()
	fields := logrus.Fields{}
	for _, metricFamily := range metricFamilies {
		fields[metricFamily.GetName()] = metricFamily.String()
	}
	return fields
}

func getSubsystem(o *Output, sample metrics.Sample) string {
	var sb strings.Builder
	for _, label := range o.config.LabelSegregation {
		if tag, exists := sample.GetTags().Get(label); exists {
			if (sb.Len() > 0) {
				sb.WriteString("_")
				sb.WriteString(tag)
			} else {
				sb.WriteString(tag)
			}
		}
	}
	return sb.String()
}