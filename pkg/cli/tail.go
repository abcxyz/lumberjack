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
	"strings"
	"time"

	logging "cloud.google.com/go/logging/apiv2"
	"cloud.google.com/go/logging/apiv2/loggingpb"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/abcxyz/lumberjack/pkg/cloudlogging"
	"github.com/abcxyz/lumberjack/pkg/validation"
	"github.com/abcxyz/pkg/cli"
	pkglogging "github.com/abcxyz/pkg/logging"
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
	StreamPull(context.Context, string, chan<- *loggingpb.LogEntry) error
}

var _ cli.Command = (*TailCommand)(nil)

// TailCommand tails and validates(optional) lumberjack logs.
type TailCommand struct {
	cli.BaseCommand

	flagScope string

	flagValidate bool

	flagMaxNum int

	flagDuration time.Duration

	flagAdditionalFilter string

	flagOverrideFilter string

	flagAdditionalCheck bool

	flagFollow bool

	// For testing only.
	testPuller logPuller
}

func (c *TailCommand) Desc() string {
	return `Tail lumberjack logs from GCP Cloud logging`
}

func (c *TailCommand) Help() string {
	return `
Usage: {{ COMMAND }} [options]

Tails and validates the latest 10 lumberjack log in the last 2 hours in the scope:

      {{ COMMAND }} -scope "projects/foo" -validate

Tails the latest 10 lumberjack log filtered by additional custom log filter:

      {{ COMMAND }} -scope "projects/foo" -additional-filter "resource.type = \"foo\""

Tails and validates (with additional check) the latest 10 lumberjack log in the last 4 hours in the scope:

      {{ COMMAND }} -scope "projects/foo" -max-num 10 -duration 4h -validate -additional-check

Streams and validates the latest lumberjack log :

      {{ COMMAND }} -scope "projects/foo" -follow -validate

Streams the latest lumberjack log filtered by additional custom log filter:

      {{ COMMAND }} -scope "projects/foo" -additional-filter "resource.type = \"foo\"" -follow
`
}

func (c *TailCommand) Flags() *cli.FlagSet {
	set := cli.NewFlagSet()

	// Command options
	f := set.NewSection("COMMAND OPTIONS")

	f.StringVar(&cli.StringVar{
		Name:    "scope",
		Aliases: []string{"s"},
		Target:  &c.flagScope,
		Example: `projects/foo`,
		Usage: `Name of the scope/parent resource from which to retrieve log ` +
			`entries, examples are: projects/[PROJECT_ID], folders/[FOLDER_ID],` +
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
		Default: 10,
		Usage: `Maximum number of most recent logs to validate ` +
			`This flag is ignored if -follow is set`,
	})

	f.DurationVar(&cli.DurationVar{
		Name:    "duration",
		Aliases: []string{"d"},
		Target:  &c.flagDuration,
		Example: "4h",
		Default: 2 * time.Hour,
		Usage: `Log filter that determines how far back to search for log entries` +
			`This flag is ignored if -follow is set`,
	})

	f.StringVar(&cli.StringVar{
		Name:    "additional-filter",
		Target:  &c.flagAdditionalFilter,
		Example: `resource.type = "gae_app" AND severity = ERROR`,
		Usage: `Log filter in addition to lumberjack log filter used to tail ` +
			`log entries, see more on ` +
			`https://cloud.google.com/logging/docs/view/logging-query-language`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "override-filter",
		Target: &c.flagOverrideFilter,
		Hidden: true,
		Usage: `Override lumberjack log filter, when it is used, it will be ` +
			`the only filter used to tail logs`,
	})

	f.BoolVar(&cli.BoolVar{
		Name:    "additional-check",
		Target:  &c.flagAdditionalCheck,
		Default: false,
		Usage: `Use it with -validate flag to validate logs tailed with ` +
			`additional lumberjack specific checks on log labels.`,
	})

	f.BoolVar(&cli.BoolVar{
		Name:    "follow",
		Aliases: []string{"f"},
		Target:  &c.flagFollow,
		Default: false,
		Usage:   "Whether to stream the logs as they show up.",
	})

	return set
}

func (c *TailCommand) Run(ctx context.Context, args []string) (rErr error) {
	f := c.Flags()
	if err := f.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}
	args = f.Args()
	if len(args) > 0 {
		return fmt.Errorf("unexpected arguments: %q", args)
	}

	if c.flagScope == "" {
		return fmt.Errorf("scope is required")
	}

	// Request with negative and greater than 1000 (log count limit) is rejected.
	if c.flagMaxNum <= 0 || c.flagMaxNum > 1000 {
		return fmt.Errorf("-max-num must be greater than 0 and less than 1000")
	}

	var extra []validation.Validator
	if c.flagAdditionalCheck {
		extra = append(extra, validation.ValidateLabels)
	}

	var puller logPuller
	if c.testPuller != nil {
		puller = c.testPuller
	} else {
		logClient, err := logging.NewClient(ctx)
		if err != nil {
			return errors.Join(rErr, fmt.Errorf("failed to create logging client: %w", err))
		}

		defer func() {
			if err := logClient.Close(); err != nil {
				rErr = errors.Join(rErr, fmt.Errorf("failed to close log client: %w", err))
			}
		}()
		puller = cloudlogging.NewPuller(ctx, logClient, c.flagScope)
	}

	if c.flagFollow {
		if err := c.stream(ctx, extra, puller); err != nil {
			rErr = errors.Join(rErr, err)
		}
	} else {
		if err := c.list(ctx, extra, puller); err != nil {
			rErr = errors.Join(rErr, err)
		}
	}
	return
}

func (c *TailCommand) list(ctx context.Context, extra []validation.Validator, puller logPuller) error {
	ls, err := puller.Pull(ctx, c.queryFilter(), c.flagMaxNum)
	if err != nil {
		return fmt.Errorf("failed to pull logs: %w", err)
	}

	if len(ls) == 0 {
		c.Outf("No logs found.")
		return nil
	}

	// Output results.
	var failCount int
	for _, l := range ls {
		js, err := protojson.Marshal(l)
		if err != nil {
			failCount++
			c.Errf("failed to marshal log to json (InsertId: %q): %w", l.GetInsertId(), err)
			continue
		}

		// Output tailed log, all spaces are stripped to reduce unit test flakiness
		// as protojson.Marshal can produce inconsistent output. See issue
		// https://github.com/golang/protobuf/issues/1121.
		c.Outf(stripSpaces(string(js)))

		// Output validation result if validation is enabled.
		if c.flagValidate {
			if err := validation.Validate(string(js), extra...); err != nil {
				failCount++
				c.Errf("failed to validate log (InsertId: %q): %w\n", l.GetInsertId(), err)
			} else {
				c.Outf("Successfully validated log (InsertId: %q)\n", l.GetInsertId())
			}
		}
	}
	if c.flagValidate {
		c.Outf("Validation failed for %d logs (out of %d)", failCount, len(ls))
	}
	return nil
}

func (c *TailCommand) stream(ctx context.Context, extra []validation.Validator, puller logPuller) error {
	// logCh is used to receive logEntriesfrom StreamPull
	logCh := make(chan *loggingpb.LogEntry)

	var failCount, totalCount int

	var perr error

	go func() {
		defer close(logCh)
		if err := puller.StreamPull(ctx, c.queryFilter(), logCh); err != nil {
			perr = fmt.Errorf("StreamPull failed: %w", err)
		}
	}()

	for l := range logCh {
		totalCount++
		js, err := protojson.Marshal(l)
		if err != nil {
			c.Errf("failed to marshal log to json (InsertId: %q): %w", l.GetInsertId(), err)
			continue
		}
		c.Outf(stripSpaces(string(js)))
		if c.flagValidate {
			if err := validation.Validate(string(js), extra...); err != nil {
				failCount++
				c.Errf("failed to validate log (InsertId: %q): %w\n", l.GetInsertId(), err)
			} else {
				c.Outf("Successfully validated log (InsertId: %q)\n", l.GetInsertId())
			}
		}
	}

	if c.flagValidate {
		c.Outf("Validation failed for %d logs (out of %d)", failCount, totalCount)
	}

	// perr will be context.Canceled if it's timed out or user pressed ctrl+c
	// in this case, the perr is logged instead of returned.
	if errors.Is(perr, context.Canceled) {
		pkglogging.FromContext(ctx).DebugContext(ctx, "stream cancelled", "error", ctx.Err())
		return nil
	}

	return perr
}

func (c *TailCommand) queryFilter() string {
	// When override filter is set, use it only to query logs.
	if c.flagOverrideFilter != "" {
		return c.flagOverrideFilter
	}

	cutoff := fmt.Sprintf("timestamp >= %q", time.Now().UTC().Add(-c.flagDuration).Format(time.RFC3339))
	f := fmt.Sprintf("%s AND %s", logType, cutoff)

	if c.flagAdditionalFilter == "" {
		return f
	}
	return fmt.Sprintf("%s AND %s", f, c.flagAdditionalFilter)
}

func stripSpaces(s string) string {
	return strings.Replace(s, " ", "", -1)
}
