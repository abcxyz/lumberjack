package v1alpha1

import (
	"bytes"
	"testing"

	"github.com/abcxyz/lumberjack/clients/go/pkg/errutil"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/viper"
)

func TestConfig(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		cfg        string
		wantConfig Config
	}{{
		name: "full config",
		cfg: `version: v1alpha1
backend:
  address: service:80
  impersonate_account: "foo@example.com"
condition:
  regex:
    principal_include: "@example.com$"
security_context:
  from_raw_jwt:
    key: x-auth
    prefix: bar
    jwks:
      endpoint: example.com/jwks
rules:
- selector: com.example.*
  directive: AUDIT
  log_type: ADMIN_ACTIVITY`,
		wantConfig: Config{
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
				FromRawJWT: &FromRawJWT{
					Key:    "x-auth",
					Prefix: "bar",
					JWKs: &JWKs{
						Endpoint: "example.com/jwks",
					},
				},
			},
			Rules: []*AuditRule{{
				Selector:  "com.example.*",
				Directive: "AUDIT",
				LogType:   "ADMIN_ACTIVITY",
			}},
		},
	}, {
		name: "minimal config",
		cfg: `version: v1alpha1
security_context:
  from_raw_jwt: {}
rules:
- selector: com.example.*
  directive: AUDIT`,
		wantConfig: Config{
			Version: "v1alpha1",
			Rules: []*AuditRule{{
				Selector:  "com.example.*",
				Directive: "AUDIT",
			}},
		},
	}}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			v := viper.New()
			v.SetConfigType("yaml")
			if err := v.ReadConfig(bytes.NewBufferString(tc.cfg)); err != nil {
				t.Fatalf("failed to load config: %v", err)
			}

			var gotConfig Config
			if err := v.Unmarshal(&gotConfig); err != nil {
				t.Fatalf("failed to unmarshal config: %v", err)
			}

			if diff := cmp.Diff(tc.wantConfig, gotConfig); diff != "" {
				t.Errorf("Config unexpected diff (-want,+got):\n%s", diff)
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
				FromRawJWT: &FromRawJWT{},
			},
			Condition: &Condition{
				Regex: &RegexCondition{},
			},
			Rules: []*AuditRule{{Selector: "*"}},
		},
	}, {
		name: "invalid version",
		cfg: &Config{
			Version: "random",
			SecurityContext: &SecurityContext{
				FromRawJWT: &FromRawJWT{},
			},
			Condition: &Condition{
				Regex: &RegexCondition{},
			},
			Rules: []*AuditRule{{Selector: "*"}},
		},
		wantErr: `unexpected Version "random"`,
	}, {
		name: "missing security context",
		cfg: &Config{
			Version: "v1alpha1",
			Condition: &Condition{
				Regex: &RegexCondition{},
			},
			Rules: []*AuditRule{{Selector: "*"}},
		},
		wantErr: "SecurityContext is nil",
	}, {
		name: "empty security context",
		cfg: &Config{
			Version:         "v1alpha1",
			SecurityContext: &SecurityContext{},
			Condition: &Condition{
				Regex: &RegexCondition{},
			},
			Rules: []*AuditRule{{Selector: "*"}},
		},
		wantErr: "one and only one SecurityContext option must be specified",
	}, {
		name: "missing rule selector",
		cfg: &Config{
			Version: "v1alpha1",
			SecurityContext: &SecurityContext{
				FromRawJWT: &FromRawJWT{},
			},
			Condition: &Condition{
				Regex: &RegexCondition{},
			},
			Rules: []*AuditRule{{}},
		},
		wantErr: "audit rule selector is empty",
	}, {
		name: "combination of errors",
		cfg: &Config{
			Version: "random",
			Rules:   []*AuditRule{{}},
		},
		wantErr: `unexpected Version "random"; SecurityContext is nil; audit rule selector is empty`,
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
		name: "default version",
		cfg:  &Config{},
		wantCfg: &Config{
			Version: "v1alpha1",
		},
	}, {
		name: "default security context",
		cfg: &Config{
			Version: "v1alpha1",
			SecurityContext: &SecurityContext{
				FromRawJWT: &FromRawJWT{},
			},
		},
		wantCfg: &Config{
			Version: "v1alpha1",
			SecurityContext: &SecurityContext{
				FromRawJWT: &FromRawJWT{
					Key:    "authorization",
					Prefix: "Bearer ",
				},
			},
		},
	}, {
		name: "default rule fields",
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
		name: "default regex condition",
		cfg: &Config{
			Version: "v1alpha1",
			Condition: &Condition{
				Regex: &RegexCondition{},
			},
		},
		wantCfg: &Config{
			Version: "v1alpha1",
			Condition: &Condition{
				Regex: &RegexCondition{
					PrincipalExclude: ".gserviceaccount.com$",
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
