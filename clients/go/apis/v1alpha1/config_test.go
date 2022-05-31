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

	"github.com/abcxyz/lumberjack/clients/go/pkg/errutil"
	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v2"
)

func TestConfig(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		cfg         string
		wantConfig  *Config
		wantLogMode AuditLogRequest_LogMode
	}{{
		name: "full_config",
		cfg: `version: v1alpha1
backend:
  remote:
    address: service:80
    impersonate_account: "foo@example.com"
condition:
  regex:
    principal_include: "@example.com$"
security_context:
  from_raw_jwt:
  - key: x-auth
    prefix: bar
    jwks:
      endpoint: example.com/jwks
rules:
- selector: com.example.*
  directive: AUDIT
  log_type: ADMIN_ACTIVITY
labels:
  mylabel1: myvalue1
  mylabel2: myvalue2
log_mode: BEST_EFFORT`,
		wantConfig: &Config{
			Version: "v1alpha1",
			Backend: &Backend{
				Remote: &Remote{
					Address:            "service:80",
					ImpersonateAccount: "foo@example.com",
				},
			},
			Condition: &Condition{
				Regex: &RegexCondition{
					PrincipalInclude: "@example.com$",
				},
			},
			SecurityContext: &SecurityContext{
				FromRawJWT: []*FromRawJWT{{
					Key:    "x-auth",
					Prefix: "bar",
					JWKs: &JWKs{
						Endpoint: "example.com/jwks",
					},
				}},
			},
			Rules: []*AuditRule{{
				Selector:  "com.example.*",
				Directive: "AUDIT",
				LogType:   "ADMIN_ACTIVITY",
			}},
			Labels: map[string]string{
				"mylabel1": "myvalue1",
				"mylabel2": "myvalue2",
			},
			LogMode: "BEST_EFFORT",
		},
		wantLogMode: AuditLogRequest_BEST_EFFORT,
	}, {
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
		tc := tc
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
	}{{
		name: "valid",
		cfg: &Config{
			Version: "v1alpha1",
			SecurityContext: &SecurityContext{
				FromRawJWT: []*FromRawJWT{{
					Key: "authorization",
				}},
			},
			Backend: &Backend{
				Remote: &Remote{
					Address: "foo",
				},
				CloudLogging: &CloudLogging{
					DefaultProject: true,
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
	}, {
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
	}, {
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
	}, {
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
	}, {
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
	}, {
		name: "combination_of_errors",
		cfg: &Config{
			Version: "random",
			Backend: &Backend{
				Remote: &Remote{Address: "fake"},
			},
			Rules: []*AuditRule{{}},
		},
		wantErr: `unexpected Version "random" want "v1alpha1"; audit rule selector is empty`,
	}, {
		name: "invalid_security_context",
		cfg: &Config{
			Version: "v1alpha1",
			SecurityContext: &SecurityContext{
				FromRawJWT: []*FromRawJWT{{
					Key: "",
				}},
			},
			Backend: &Backend{
				Remote: &Remote{
					Address: "foo",
				},
			},
		},
		wantErr: `FromRawJWT[0]: key must be specified`,
	}, {
		name: "invalid_log_mode",
		cfg: &Config{
			Version: "v1alpha1",
			Backend: &Backend{
				Remote: &Remote{
					Address: "foo",
				},
			},
			LogMode: "random",
		},
		wantErr: `invalid LogMode "random"`,
	}, {
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
	}, {
		name: "invalid_backend_cloudlogging_no_default",
		cfg: &Config{
			Version: "v1alpha1",
			SecurityContext: &SecurityContext{
				FromRawJWT: []*FromRawJWT{{
					Key: "authorization",
				}},
			},
			Backend: &Backend{
				CloudLogging: &CloudLogging{
					DefaultProject: false,
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
		wantErr: `backend cloudlogging no project or using default project is set`,
	}}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.cfg.Validate()
			if diff := errutil.DiffSubstring(err, tc.wantErr); diff != "" {
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
		cfg:  &Config{},
		wantCfg: &Config{
			Version: "v1alpha1",
		},
	}, {
		name: "default_rule_fields",
		cfg: &Config{
			Version: "v1alpha1",
			Rules: []*AuditRule{{
				Selector: "*",
			}},
		},
		wantCfg: &Config{
			Version: "v1alpha1",
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
			Backend: &Backend{
				CloudLogging: &CloudLogging{},
			},
		},
		wantCfg: &Config{
			Version: "v1alpha1",
			Backend: &Backend{
				CloudLogging: &CloudLogging{
					DefaultProject: true,
				},
			},
		},
	}}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.cfg.SetDefault()
			if diff := cmp.Diff(tc.wantCfg, tc.cfg); diff != "" {
				t.Errorf("SetDefault() unexpected diff (-want,+got):\n%s", diff)
			}
		})
	}
}
