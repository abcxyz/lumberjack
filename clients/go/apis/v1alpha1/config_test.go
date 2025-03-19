// Copyright 2022 Lumberjack authors (see AUTHORS file)
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

package v1alpha1

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v2"

	pkgtestutil "github.com/abcxyz/pkg/testutil"
)

func TestConfig(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		cfg         string
		wantConfig  *Config
		wantLogMode AuditLogRequest_LogMode
	}{{
		name: "minimal_config",
		cfg: `version: v1alpha1
security_context:
  from_raw_jwt:
  - {}
rules:
- selector: com.example.*
  directive: AUDIT`,
		wantConfig: &Config{
			Version: "v1alpha1",
			Rules: []*AuditRule{{
				Selector:  "com.example.*",
				Directive: "AUDIT",
			}},
			SecurityContext: &SecurityContext{FromRawJWT: []*FromRawJWT{{}}},
		},
		wantLogMode: AuditLogRequest_LOG_MODE_UNSPECIFIED,
	}, {
		name: "config_with_just_version",
		cfg:  `version: v1alpha1`,
		wantConfig: &Config{
			Version: "v1alpha1",
		},
		wantLogMode: AuditLogRequest_LOG_MODE_UNSPECIFIED,
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			content := bytes.NewBufferString(tc.cfg).Bytes()
			gotConfig := &Config{}
			if err := yaml.Unmarshal(content, gotConfig); err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.wantConfig, gotConfig); diff != "" {
				t.Errorf("Config unexpected diff (-want,+got):\n%s", diff)
			}

			if got, want := gotConfig.GetLogMode(), tc.wantLogMode; got != want {
				t.Errorf("wanted log mode %v but got %v", want, got)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		cfg     *Config
		wantErr string
	}{
		{
			name: "invalid_version",
			cfg: &Config{
				Version: "random",
				SecurityContext: &SecurityContext{
					FromRawJWT: []*FromRawJWT{},
				},
				Condition: &Condition{
					Regex: &RegexCondition{},
				},
				Rules: []*AuditRule{{Selector: "*"}},
			},
			wantErr: `unexpected Version "random" want "v1alpha1"`,
		},
		{
			name: "missing_rule_selector",
			cfg: &Config{
				Version: "v1alpha1",
				SecurityContext: &SecurityContext{
					FromRawJWT: []*FromRawJWT{},
				},
				Condition: &Condition{
					Regex: &RegexCondition{},
				},
				Rules: []*AuditRule{{}},
			},
			wantErr: "audit rule selector is empty",
		},
		{
			name: "invalid_rule_directive",
			cfg: &Config{
				Version: "v1alpha1",
				SecurityContext: &SecurityContext{
					FromRawJWT: []*FromRawJWT{},
				},
				Condition: &Condition{
					Regex: &RegexCondition{},
				},
				Rules: []*AuditRule{{
					Selector:  "*",
					Directive: "random",
					LogType:   "DATA_ACCESS",
				}},
			},
			wantErr: `unexpected rule.Directive "random" want one of ["AUDIT", "AUDIT_REQUEST_ONLY"`,
		},
		{
			name: "invalid_rule_log_type",
			cfg: &Config{
				Version: "v1alpha1",
				SecurityContext: &SecurityContext{
					FromRawJWT: []*FromRawJWT{},
				},
				Condition: &Condition{
					Regex: &RegexCondition{},
				},
				Rules: []*AuditRule{{
					Selector:  "*",
					Directive: "AUDIT_REQUEST_ONLY",
					LogType:   "random",
				}},
			},
			wantErr: `unexpected rule.LogType "random" want one of ["ADMIN_ACTIVITY", "DATA_ACCESS"]`,
		},
		{
			name: "invalid_backend_cloudlogging_use_default",
			cfg: &Config{
				Version: "v1alpha1",
				SecurityContext: &SecurityContext{
					FromRawJWT: []*FromRawJWT{{
						Key: "authorization",
					}},
				},
				Backend: &Backend{
					CloudLogging: &CloudLogging{
						DefaultProject: true,
						Project:        "my-other-proj",
					},
				},
				Condition: &Condition{
					Regex: &RegexCondition{},
				},
				Rules: []*AuditRule{{
					Selector:  "*",
					Directive: "AUDIT_REQUEST_ONLY",
					LogType:   "DATA_ACCESS",
				}},
			},
			wantErr: `backend cloudlogging project is set while using default project`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.cfg.Validate()
			if diff := pkgtestutil.DiffErrString(err, tc.wantErr); diff != "" {
				t.Errorf("Validate() got unexpected error: %s", diff)
			}
		})
	}
}

func TestSetDefault(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		cfg     *Config
		wantCfg *Config
	}{{
		name: "default_version",
		cfg: &Config{
			LogMode: "BEST_EFFORT",
		},
		wantCfg: &Config{
			Version: "v1alpha1",
			LogMode: "BEST_EFFORT",
		},
	}, {
		name: "default_rule_fields",
		cfg: &Config{
			Version: "v1alpha1",
			LogMode: "BEST_EFFORT",
			Rules: []*AuditRule{{
				Selector: "*",
			}},
		},
		wantCfg: &Config{
			Version: "v1alpha1",
			LogMode: "BEST_EFFORT",
			Rules: []*AuditRule{{
				Selector:  "*",
				Directive: "AUDIT",
				LogType:   "DATA_ACCESS",
			}},
		},
	}, {
		name: "default_backend_cloudlogging",
		cfg: &Config{
			Version: "v1alpha1",
			LogMode: "BEST_EFFORT",
			Backend: &Backend{
				CloudLogging: &CloudLogging{},
			},
		},
		wantCfg: &Config{
			Version: "v1alpha1",
			LogMode: "BEST_EFFORT",
			Backend: &Backend{
				CloudLogging: &CloudLogging{
					DefaultProject: true,
				},
			},
		},
	}, {
		name: "default_fail_close_log_mode",
		cfg: &Config{
			Version: "v1alpha1",
		},
		wantCfg: &Config{
			Version: "v1alpha1",
			LogMode: AuditLogRequest_FAIL_CLOSE.String(),
		},
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.cfg.SetDefault()
			if diff := cmp.Diff(tc.wantCfg, tc.cfg); diff != "" {
				t.Errorf("SetDefault() unexpected diff (-want,+got):\n%s", diff)
			}
		})
	}
}

func TestGetLogMode(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		cfg  *Config
		want AuditLogRequest_LogMode
	}{
		{
			name: "unspecified_if_missing",
			cfg:  &Config{},
			want: AuditLogRequest_LOG_MODE_UNSPECIFIED,
		},
		{
			name: "fail_close_upper_case",
			cfg: &Config{
				LogMode: "FAIL_CLOSE",
			},
			want: AuditLogRequest_FAIL_CLOSE,
		},
		{
			name: "fail_close_lower_case",
			cfg: &Config{
				LogMode: "fail_close",
			},
			want: AuditLogRequest_FAIL_CLOSE,
		},
		{
			name: "best_effort_upper_case",
			cfg: &Config{
				LogMode: "BEST_EFFORT",
			},
			want: AuditLogRequest_BEST_EFFORT,
		},
		{
			name: "best_effort_lower_case",
			cfg: &Config{
				LogMode: "best_effort",
			},
			want: AuditLogRequest_BEST_EFFORT,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got, want := tc.cfg.GetLogMode(), tc.want; got != want {
				t.Errorf("log mode got %v want %v", got, want)
			}
		})
	}
}
