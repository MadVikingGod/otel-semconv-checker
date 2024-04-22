// SPDX-License-Identifier: Apache-2.0

package sccconnector

import (
	"errors"
	"regexp"

	"github.com/madvikinggod/otel-semconv-checker/pkg/match"
	"go.opentelemetry.io/collector/component"
)

type Config struct {
	Trace []match.Match `mapstructure:"trace"`
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

func createDefaultConfig() component.Config {
	return &Config{
		Trace: []match.Match{
			{
				Match: "http.server.*",
				Groups: []string{
					"trace.http.server",
				},
				ReportAdditional: true,
			},
			{
				Match: "http.client.*",
				Groups: []string{
					"trace.http.client",
				},
				ReportAdditional: true,
			},
		},
	}
}
