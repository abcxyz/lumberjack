package v1alpha1

import (
	"bytes"
	"testing"

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
