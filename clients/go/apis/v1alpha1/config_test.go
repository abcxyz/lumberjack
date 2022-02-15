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
log_mode: BEST_EFFORT`,
		wantConfig: &Config{
			Version: "v1alpha1",
			Backend: &Backend{
				Address:            "service:80",
				ImpersonateAccount: "foo@example.com",
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
			Backend: &Backend{Address: "foo"},
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
			Rules:   []*AuditRule{{}},
		},
		wantErr: `unexpected Version "random" want "v1alpha1"; backend is nil; audit rule selector is empty`,
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
