// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package helper // import "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"

import (
	"context"

	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/entry"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/errors"
)

// NewParserConfig creates a new parser config with default values
func NewParserConfig(operatorID, operatorType string) ParserConfig {
	return ParserConfig{
		TransformerConfig: NewTransformerConfig(operatorID, operatorType),
		ParseFrom:         entry.NewBodyField(),
		ParseTo:           entry.NewAttributeField(),
	}
}

// ParserConfig provides the basic implementation of a parser config.
type ParserConfig struct {
	TransformerConfig `mapstructure:",squash" yaml:",inline"`

	ParseFrom       entry.Field      `mapstructure:"parse_from"          json:"parse_from"          yaml:"parse_from"`
	ParseTo         entry.Field      `mapstructure:"parse_to"            json:"parse_to"            yaml:"parse_to"`
	TimeParser      *TimeParser      `mapstructure:"timestamp,omitempty" json:"timestamp,omitempty" yaml:"timestamp,omitempty"`
	Config          *Config          `mapstructure:"severity,omitempty"  json:"severity,omitempty"  yaml:"severity,omitempty"`
	TraceParser     *TraceParser     `mapstructure:"trace,omitempty"     json:"trace,omitempty"     yaml:"trace,omitempty"`
	ScopeNameParser *ScopeNameParser `mapstructure:"scope_name,omitempty"     json:"scope_name,omitempty"     yaml:"scope_name,omitempty"`
}

// Build will build a parser operator.
func (c ParserConfig) Build(logger *zap.SugaredLogger) (ParserOperator, error) {
	transformerOperator, err := c.TransformerConfig.Build(logger)
	if err != nil {
		return ParserOperator{}, err
	}

	parserOperator := ParserOperator{
		TransformerOperator: transformerOperator,
		ParseFrom:           c.ParseFrom,
		ParseTo:             c.ParseTo,
	}

	if c.TimeParser != nil {
		if err := c.TimeParser.Validate(); err != nil {
			return ParserOperator{}, err
		}
		parserOperator.TimeParser = c.TimeParser
	}

	if c.Config != nil {
		severityParser, err := c.Config.Build(logger)
		if err != nil {
			return ParserOperator{}, err
		}
		parserOperator.SeverityParser = &severityParser
	}

	if c.TraceParser != nil {
		if err := c.TraceParser.Validate(); err != nil {
			return ParserOperator{}, err
		}
		parserOperator.TraceParser = c.TraceParser
	}

	if c.ScopeNameParser != nil {
		parserOperator.ScopeNameParser = c.ScopeNameParser
	}

	return parserOperator, nil
}

// ParserOperator provides a basic implementation of a parser operator.
type ParserOperator struct {
	TransformerOperator
	ParseFrom       entry.Field
	ParseTo         entry.Field
	TimeParser      *TimeParser
	SeverityParser  *SeverityParser
	TraceParser     *TraceParser
	ScopeNameParser *ScopeNameParser
}

// ProcessWith will run ParseWith on the entry, then forward the entry on to the next operators.
func (p *ParserOperator) ProcessWith(ctx context.Context, entry *entry.Entry, parse ParseFunction) error {
	return p.ProcessWithCallback(ctx, entry, parse, nil)
}

func (p *ParserOperator) ProcessWithCallback(ctx context.Context, entry *entry.Entry, parse ParseFunction, cb func(*entry.Entry) error) error {
	// Short circuit if the "if" condition does not match
	skip, err := p.Skip(ctx, entry)
	if err != nil {
		return p.HandleEntryError(ctx, entry, err)
	}
	if skip {
		p.Write(ctx, entry)
		return nil
	}

	if err = p.ParseWith(ctx, entry, parse); err != nil {
		return err
	}
	if cb != nil {
		err = cb(entry)
		if err != nil {
			return err
		}
	}

	p.Write(ctx, entry)
	return nil
}

// ParseWith will process an entry's field with a parser function.
func (p *ParserOperator) ParseWith(ctx context.Context, entry *entry.Entry, parse ParseFunction) error {
	value, ok := entry.Get(p.ParseFrom)
	if !ok {
		err := errors.NewError(
			"Entry is missing the expected parse_from field.",
			"Ensure that all incoming entries contain the parse_from field.",
			"parse_from", p.ParseFrom.String(),
		)
		return p.HandleEntryError(ctx, entry, err)
	}

	newValue, err := parse(value)
	if err != nil {
		return p.HandleEntryError(ctx, entry, err)
	}

	if err := entry.Set(p.ParseTo, newValue); err != nil {
		return p.HandleEntryError(ctx, entry, errors.Wrap(err, "set parse_to"))
	}

	var timeParseErr error
	if p.TimeParser != nil {
		timeParseErr = p.TimeParser.Parse(entry)
	}

	var severityParseErr error
	if p.SeverityParser != nil {
		severityParseErr = p.SeverityParser.Parse(entry)
	}

	var traceParseErr error
	if p.TraceParser != nil {
		traceParseErr = p.TraceParser.Parse(entry)
	}

	var scopeNameParserErr error
	if p.ScopeNameParser != nil {
		scopeNameParserErr = p.ScopeNameParser.Parse(entry)
	}

	// Handle time or severity parsing errors after attempting to parse both
	if timeParseErr != nil {
		return p.HandleEntryError(ctx, entry, errors.Wrap(timeParseErr, "time parser"))
	}
	if severityParseErr != nil {
		return p.HandleEntryError(ctx, entry, errors.Wrap(severityParseErr, "severity parser"))
	}
	if traceParseErr != nil {
		return p.HandleEntryError(ctx, entry, errors.Wrap(traceParseErr, "trace parser"))
	}
	if scopeNameParserErr != nil {
		return p.HandleEntryError(ctx, entry, errors.Wrap(scopeNameParserErr, "scope_name parser"))
	}
	return nil
}

// ParseFunction is function that parses a raw value.
type ParseFunction = func(interface{}) (interface{}, error)
