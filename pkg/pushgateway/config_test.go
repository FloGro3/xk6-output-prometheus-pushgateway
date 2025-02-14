package pushgateway_test

import (
	"testing"
	"slices"

	"github.com/FloGro3/xk6-output-prometheus-pushgateway/pkg/pushgateway"
	"github.com/sirupsen/logrus"
	"go.k6.io/k6/output"
)

func TestConfigLabels(t *testing.T) {
	p := output.Params{
		Environment: map[string]string{},
		Logger:      logrus.New(),
	}
	cfg, err := pushgateway.NewConfig(p)
	if err != nil {
		t.Errorf("Unable to create a new config, error: %v", err)
	}
	if len(cfg.LabelSegregation) != 1 {
		t.Errorf("Unexpecten labels value %+v", cfg)
	}

	p.Environment["K6_LABEL_SEGREGATION"] = "PROD, APP"
	cfg, err = pushgateway.NewConfig(p)
	if err != nil {
		t.Errorf("Unable to create a new config, error: %v", err)
	}
	if len(cfg.LabelSegregation) == 3 &&
		slices.Contains(cfg.LabelSegregation, "APP") &&	
		slices.Contains(cfg.LabelSegregation, "PROD") {
		t.Errorf("Unexpecten labels value %+v", cfg)
	}
}
