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
	"errors"
	"fmt"
	"time"

	"cloud.google.com/go/logging/apiv2/loggingpb"
	"github.com/abcxyz/lumberjack/pkg/cloudlogging"
	"github.com/abcxyz/lumberjack/pkg/validation"
	"github.com/abcxyz/pkg/cli"
	"google.golang.org/protobuf/encoding/protojson"

	logging "cloud.google.com/go/logging/apiv2"
)

// Lumberjack specific log types.
const logType = `LOG_ID("audit.abcxyz/unspecified") OR ` +
	`LOG_ID("audit.abcxyz/activity") OR ` +
	`LOG_ID("audit.abcxyz/data_access") OR ` +
	`LOG_ID("audit.abcxyz/consent") OR ` +
	`LOG_ID("audit.abcxyz/system_event")`

// logPuller interface that pulls log entries from cloud logging.
type logPuller interface {
	Pull(context.Context, string, int) ([]*loggingpb.LogEntry, error)
}

var _ cli.Command = (*TailCommand)(nil)

// TailCommand tails and validates(optional) lumberjack logs.
type TailCommand struct {
	cli.BaseCommand

	flagResource string

	flagValidate bool

	flagMaxNum int

	flagDuration time.Duration

	flagCustomQuery string

	flagRemoveLumberjackLogType bool

	flagValidateWithAdditionalCheck bool

	// For testing only.
	testPuller logPuller
}

func (c *TailCommand) Desc() string {
	return `Tail lumberjack logs from Cloud logging and validate them when validation enabled`
}

func (c *TailCommand) Help() string {
	return `
Usage: {{ COMMAND }} [options]

Tails and validates the latest lumberjack log in the last 24 hours from resource:

      {{ COMMAND }} -resource "project/foo" -validate

Tails the latest lumberjack log filtered by additional custom query:

      {{ COMMAND }} -resource "project/foo" -query "resource.type = \"foo\""

Tails and validates (with additional check) the latest 10 lumberjack log in the last 2 hours from resource:

      {{ COMMAND }} -resource "project/foo" -max-num 10 -duration 2h -validate-with-additional-check

Pulls and validates the latest non-lumberjack log type log:

      {{ COMMAND }} -resource "project/foo" -remove-lumberjack-log-type
`
}

func (c *TailCommand) Flags() *cli.FlagSet {
	set := cli.NewFlagSet()

	// Command options
	f := set.NewSection("COMMAND OPTIONS")

	f.StringVar(&cli.StringVar{
		Name:    "resource",
		Aliases: []string{"r"},
		Target:  &c.flagResource,
		Example: `projects/foo`,
		Usage: `Name of the parent resource from which to retrieve log entries,` +
			`examples are: projects/[PROJECT_ID], folders/[FOLDER_ID],` +
			`organizations/[ORGANIZATION_ID], billingAccounts/[BILLING_ACCOUNT_ID]`,
	})

	f.BoolVar(&cli.BoolVar{
		Name:    "validate",
		Aliases: []string{"v"},
		Target:  &c.flagValidate,
		Default: false,
		Usage:   `Turn on for lumberjack log validation`,
	})

	f.IntVar(&cli.IntVar{
		Name:    "max-num",
		Aliases: []string{"n"},
		Target:  &c.flagMaxNum,
		Default: 1,
		Usage:   `Maximum number of most recent logs to validate, default is 1`,
	})

	f.DurationVar(&cli.DurationVar{
		Name:    "duration",
		Aliases: []string{"d"},
		Target:  &c.flagDuration,
		Example: "2h",
		Default: 24 * time.Hour,
		Usage:   `How far back to search for log entries, default is 24 hours`,
	})

	f.StringVar(&cli.StringVar{
		Name:    "query",
		Target:  &c.flagCustomQuery,
		Example: `resource.type = "gae_app" AND severity = ERROR`,
		Usage: `Optional custom log queries, see more on ` +
			`https://cloud.google.com/logging/docs/view/logging-query-language`,
	})

	f.BoolVar(&cli.BoolVar{
		Name:    "remove-lumberjack-log-type",
		Target:  &c.flagRemoveLumberjackLogType,
		Default: false,
		Usage:   `Turn on to remove lumberjack log type log filter`,
	})

	f.BoolVar(&cli.BoolVar{
		Name:    "validate-with-additional-check",
		Target:  &c.flagValidateWithAdditionalCheck,
		Default: false,
		Usage: `Turn on for lumberjack log validation with additional ` +
			`lumberjack specific checks on log labels.`,
	})

	return set
}

func (c *TailCommand) Run(ctx context.Context, args []string) error {
	f := c.Flags()
	if err := f.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}
	args = f.Args()
	if len(args) > 0 {
		return fmt.Errorf("unexpected arguments: %q", args)
	}

	if c.flagResource == "" {
		return fmt.Errorf("resource is required")
	}

	// Request with negative and greater than 1000 (log count limit) is rejected.
	if c.flagMaxNum <= 0 || c.flagMaxNum > 1000 {
		return fmt.Errorf("max number must be greater than 0 and less than 1000")
	}

	// Tail logs.
	ls, err := c.tail(ctx)
	if err != nil {
		return err
	}
	if len(ls) == 0 {
		c.Outf("Log not found")
		return nil
	}

	// Output results.
	var extra []validation.Validator
	if c.flagValidateWithAdditionalCheck {
		extra = append(extra, validation.ValidateLabels)
	}
	var retErr error
	for _, l := range ls {
		js, err := protojson.Marshal(l)
		if err != nil {
			retErr = errors.Join(retErr, fmt.Errorf("failed to marshal log to json (InsertId: %q): %w", l.InsertId, err))
			continue
		}

		// Output log entry in JSON format if validation is not enabled.
		if !c.flagValidate && !c.flagValidateWithAdditionalCheck {
			c.Outf(string(js))
			continue
		}

		// Output validation result if validation is enabled.
		if err := validation.Validate(string(js), extra...); err != nil {
			retErr = errors.Join(retErr, fmt.Errorf("failed to validate log (InsertId: %q): %w", l.InsertId, err))
		} else {
			c.Outf("Successfully validated log (InsertId: %q)", l.InsertId)
		}
	}

	return retErr
}

func (c *TailCommand) tail(ctx context.Context) ([]*loggingpb.LogEntry, error) {
	var p logPuller
	if c.testPuller != nil {
		p = c.testPuller
	} else {
		logClient, err := logging.NewClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create logging client: %w", err)
		}
		p = cloudlogging.NewPuller(ctx, logClient, c.flagResource)
	}

	ls, err := p.Pull(ctx, c.getFilter(), c.flagMaxNum)
	if err != nil {
		return nil, fmt.Errorf("failed to pull logs: %w", err)
	}

	return ls, nil
}

func (c *TailCommand) getFilter() string {
	cutoff := fmt.Sprintf("timestamp >= %q", time.Now().UTC().Add(-c.flagDuration).Format(time.RFC3339))

	var f string
	if c.flagRemoveLumberjackLogType {
		f = cutoff
	} else {
		f = fmt.Sprintf("%s AND %s", logType, cutoff)
	}

	if c.flagCustomQuery == "" {
		return f
	}
	return fmt.Sprintf("%s AND %s", f, c.flagCustomQuery)
}
