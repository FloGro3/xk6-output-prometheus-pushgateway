package pushgateway

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"

	"go.k6.io/k6/output"
)

const (
	defaultPushGWUrl    		= "http://localhost:9091"
	defaultPushInterval 		= 5 * time.Second
	defaultJobName      		= "k6_load_testing"
	defaultNamespace    		= ""
)

var (
	defaultLabelSegregation 	= []string{"scenario"}
)

// Config is the config for the template collector
type Config struct {
	JobName      		string
	PushGWUrl    		string
	PushInterval 		time.Duration
	LabelSegregation 	[]string

	// Used to prefix all tags with a custom namespace
	Namespace string
}

// NewConfig creates a new Config instance from the provided output.Params
func NewConfig(p output.Params) (Config, error) {
	cfg := Config{
		JobName:      		defaultJobName,
		PushGWUrl:    		defaultPushGWUrl,
		PushInterval: 		defaultPushInterval,
		Namespace:    		defaultNamespace,
		LabelSegregation: 	defaultLabelSegregation,
	}

	if val, ok := p.ScriptOptions.External["pushgateway"]; ok {
		err := json.Unmarshal(val, &cfg.LabelSegregation)
		if err != nil {
			j, err := json.Marshal(&val)
			if err != nil {
				return cfg, errors.Wrap(err, fmt.Sprintf(
					"unable to get labels for JSON options.ext.pushgateway dictionary %s", string(j)))
			} else {
				return cfg, errors.Wrap(err, "unable to get labels for JSON options.ext.pushgateway dictionary")
			}

		}
		p.Logger.Debugf("Pushgateway labels from JSON options.ext.pushgateway dictionary %+v", cfg.LabelSegregation)
	}

	for k, v := range p.Environment {
		switch k {
		case "K6_PUSH_INTERVAL":
			var err error
			cfg.PushInterval, err = time.ParseDuration(v)
			if err != nil {
				return cfg, fmt.Errorf("error parsing environment variable 'K6_TEMPLATE_PUSH_INTERVAL': %w", err)
			}
		case "K6_PUSHGATEWAY_URL":
			cfg.PushGWUrl = v
		case "K6_PUSHGATEWAY_NAMESPACE":
			cfg.Namespace = v
		case "K6_JOB_NAME":
			cfg.JobName = v
		case "K6_LABEL_SEGREGATION":
			trimmedValue := strings.TrimSpace(v)
			parts := strings.Split(trimmedValue, ",")
			for _, part := range parts {
				cfg.LabelSegregation = append(cfg.LabelSegregation, strings.ToLower(part))
			}
		}
	}
	p.Logger.Debugf("Pushgateway labels %+v", cfg.LabelSegregation)
	return cfg, nil
}
