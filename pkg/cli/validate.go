// Copyright 2023 The Authors (see AUTHORS file)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cli

import (
	"context"
	"fmt"

	"github.com/abcxyz/lumberjack/pkg/validation"
	"github.com/abcxyz/pkg/cli"
)

var _ cli.Command = (*ValidateCommand)(nil)

// ValidateCommand validates lumberjack logs.
type ValidateCommand struct {
	cli.BaseCommand

	flagLogEntry string

	flagAdditionalCheck bool
}

func (c *ValidateCommand) Desc() string {
	return `Validate lumberjack log`
}

func (c *ValidateCommand) Help() string {
	return `
Usage: {{ COMMAND }} [options]

Validate lumberjack log:

      {{ COMMAND }} -log-entry "{\"foo\": \"bar\"}"

Validate the lumberjack log read from pipe:

      cat log.text | {{ COMMAND }} -log-entry -
`
}

func (c *ValidateCommand) Flags() *cli.FlagSet {
	set := cli.NewFlagSet()

	// Command options
	f := set.NewSection("COMMAND OPTIONS")

	f.StringVar(&cli.StringVar{
		Name:    "log-entry",
		Aliases: []string{"l"},
		Target:  &c.flagLogEntry,
		Example: `{"foo":"bar"}`,
		Usage: `The lumberjack/data access log, in JSON format. Set the value to` +
			` "-" to read from stdin, it stops reading when it reaches end of file`,
	})

	f.BoolVar(&cli.BoolVar{
		Name:    "additional-check",
		Target:  &c.flagAdditionalCheck,
		Default: false,
		Usage:   `Turn on for additional lumberjack specific checks on log labels.`,
	})

	return set
}

func (c *ValidateCommand) Run(ctx context.Context, args []string) error {
	f := c.Flags()
	if err := f.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}
	args = f.Args()
	if len(args) > 0 {
		return fmt.Errorf("unexpected arguments: %q", args)
	}

	if c.flagLogEntry == "" {
		return fmt.Errorf("log is required")
	}

	if c.flagLogEntry == "-" {
		// Read log from stdin until it encounters an EOF.
		log, err := c.PromptAll(ctx, "Enter log: ")
		if err != nil {
			return fmt.Errorf("failed to get log from prompt: %w", err)
		}
		c.flagLogEntry = log
	}

	var extra []validation.Validator
	if c.flagAdditionalCheck {
		extra = append(extra, validation.ValidateLabels)
	}
	if err := validation.Validate(c.flagLogEntry, extra...); err != nil {
		return fmt.Errorf("failed to validate log: %w", err)
	}
	c.Outf("Successfully validated log")

	return nil
}
