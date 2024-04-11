package sccconnector

import (
	"errors"
	"regexp"

	"github.com/madvikinggod/otel-semconv-checker/pkg/match"
)

type Config struct {
	Trace []match.Match
	// Metrics         []match.Match
	// Log             []match.Match
	ReportUnmatched bool `mapstructure:"report_unmatched"`
}

func (c *Config) Validate() error {
	errs := []error{}
	for _, m := range c.Trace {
		_, err := regexp.Compile(m.Match)
		if err != nil {
			errs = append(errs, err)
		}
	}
	// for _, m := range c.Metrics {
	// 	_, err := regexp.Compile(m.Match)
	// 	if err != nil {
	// 		errs = append(errs, err)
	// 	}
	// }
	// for _, m := range c.Log {
	// 	_, err := regexp.Compile(m.Match)
	// 	if err != nil {
	// 		errs = append(errs, err)
	// 	}
	// }

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
