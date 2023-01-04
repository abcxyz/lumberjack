// Copyright 2021 Lumberjack authors (see AUTHORS file)
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

// Package filtering provides a processor to filter audit log requests.
package filtering

import (
	"context"
	"fmt"
	"regexp"

	api "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
	"github.com/abcxyz/lumberjack/clients/go/pkg/audit"
)

// PrincipalEmailMatcher applies regexp filters on AuditLogRequest
// field `Payload.AuthenticationInfo.PrincipalEmail`.
type PrincipalEmailMatcher struct {
	includes []*regexp.Regexp
	excludes []*regexp.Regexp
}

// An Option is a configuration Option for a PrincipalEmailMatcher.
type Option func(m *PrincipalEmailMatcher) error

// WithIncludes adds include filters by compiling strings into regular expressions.
// When an audit log request has a principal email that matches one of the regular
// expressions, the audit log request is allowed for further processing. Passing an
// empty string in `includes` is a noop.
func WithIncludes(includes ...string) Option {
	return func(m *PrincipalEmailMatcher) error {
		for _, i := range includes {
			if i == "" {
				continue
			}
			r, err := regexp.Compile(i)
			if err != nil {
				return fmt.Errorf("failed to compile include string %q as a regular expression: %w", i, err)
			}
			m.includes = append(m.includes, r)
		}
		return nil
	}
}

// WithExcludes adds include filters by compiling strings into regular expressions.
// When an audit log request has a principal email that matches one of the regular
// expressions, the audit log request is dropped by the client. Passing an empty
// string in `excludes` is a noop.
func WithExcludes(excludes ...string) Option {
	return func(m *PrincipalEmailMatcher) error {
		for _, e := range excludes {
			if e == "" {
				continue
			}
			r, err := regexp.Compile(e)
			if err != nil {
				return fmt.Errorf("failed to compile exclude string %q as a regular expression: %w", e, err)
			}
			m.excludes = append(m.excludes, r)
		}
		return nil
	}
}

// NewPrincipalEmailMatcher creates a PrincipalEmailMatcher with the given options.
func NewPrincipalEmailMatcher(opts ...Option) (*PrincipalEmailMatcher, error) {
	m := &PrincipalEmailMatcher{}
	for _, o := range opts {
		if err := o(m); err != nil {
			return nil, fmt.Errorf("failed to apply NewPrincipalEmailMatcher options: %w", err)
		}
	}
	return m, nil
}

// Process with receiver PrincipalEmailMatcher filters log requests when the
// principal email matches an `include` or `exclude` regular expression. We use
// the following filtering logic:
//
//  1. If include == nil and exclude == nil, we pass the request.
//
//  2. If include != nil and exclude == nil, we only pass the request when the
//     principal matches include.
//
//  3. If include == nil and exclude != nil, we only drop the request when the
//     principal matches exclude.
//
//  4. If include != nil and exclude != nil, we drop the request when the
//     principal doesn't match include and matches exclude.
func (p *PrincipalEmailMatcher) Process(_ context.Context, logReq *api.AuditLogRequest) error {
	if len(p.includes) == 0 && len(p.excludes) == 0 {
		return nil
	}

	if logReq.Payload == nil || logReq.Payload.AuthenticationInfo == nil {
		return fmt.Errorf("request.Payload.AuthenticationInfo is missing to check principal email: %w", audit.ErrInvalidRequest)
	}

	for _, r := range p.includes {
		if r.MatchString(logReq.Payload.AuthenticationInfo.PrincipalEmail) {
			return nil
		}
	}
	if len(p.excludes) == 0 {
		// Here, len(p.include) != nil and there was no match in the includes.
		// We drop the request because it was not explicitly included.
		return fmt.Errorf("request.Payload.AuthenticationInfo.PrincipalEmail not included in %q: %w", p.includes, audit.ErrFailedPrecondition)
	}

	for _, r := range p.excludes {
		if r.MatchString(logReq.Payload.AuthenticationInfo.PrincipalEmail) {
			// When explicitly excluded, drop the request.
			return fmt.Errorf("request.Payload.AuthenticationInfo.PrincipalEmail matches exclude regexp %q: %w", r, audit.ErrFailedPrecondition)
		}
	}
	// Otherwise, pass the request.
	return nil
}
