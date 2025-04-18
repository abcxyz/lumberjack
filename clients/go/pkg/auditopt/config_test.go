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

package auditopt

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sethvargo/go-envconfig"
	capi "google.golang.org/genproto/googleapis/cloud/audit"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/testing/protocmp"

	api "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
	"github.com/abcxyz/lumberjack/clients/go/pkg/audit"
	"github.com/abcxyz/lumberjack/clients/go/pkg/testutil"
	"github.com/abcxyz/pkg/logging"
	pkgtestutil "github.com/abcxyz/pkg/testutil"
)

type fakeServer struct {
	api.UnimplementedAuditLogAgentServer
	gotReq *api.AuditLogRequest
}

func (s *fakeServer) ProcessLog(_ context.Context, logReq *api.AuditLogRequest) (*api.AuditLogResponse, error) {
	s.gotReq = logReq
	return &api.AuditLogResponse{Result: logReq}, nil
}

func TestFromConfigFile(t *testing.T) {
	t.Parallel()

	ctx := logging.WithLogger(t.Context(), logging.TestLogger(t))

	configFileContentByName := map[string]string{
		"valid.yaml": `
version: v1alpha1
condition:
  regex:
    principal_exclude: user@example.com$
backend:
  remote:
    # we set the backend address as env var below
    insecure_enabled: true
`,
		"invalid.yaml": `bananas`,
	}

	// Set up config files.
	dir := t.TempDir()
	for name, content := range configFileContentByName {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
		})
	}

	cases := []struct {
		name          string
		envs          map[string]string
		path          string
		req           *api.AuditLogRequest
		wantReq       *api.AuditLogRequest
		wantErrSubstr string
	}{
		{
			name: "use_config_file_when_provided",
			path: path.Join(dir, "valid.yaml"),
			req:  testutil.NewRequest(testutil.WithPrincipal("abc@project.iam.gserviceaccount.com")),
			wantReq: testutil.NewRequest(
				testutil.WithPrincipal("abc@project.iam.gserviceaccount.com"),
				testutil.WithMode(api.AuditLogRequest_FAIL_CLOSE)),
		},
		{
			name: "use_env_var_when_config_file_not_found",
			path: path.Join(dir, "inexistent.yaml"),
			envs: map[string]string{
				"AUDIT_CLIENT_CONDITION_REGEX_PRINCIPAL_EXCLUDE":  "user@example.com$",
				"AUDIT_CLIENT_BACKEND_REMOTE_INSECURE_ENABLED":    "true",
				"AUDIT_CLIENT_BACKEND_REMOTE_IMPERSONATE_ACCOUNT": "example@test.iam.gserviceaccount.com",
				"AUDIT_CLIENT_LOG_MODE":                           "BEST_EFFORT",
			},
			req: testutil.NewRequest(testutil.WithPrincipal("abc@project.iam.gserviceaccount.com")),
			wantReq: testutil.NewRequest(
				testutil.WithPrincipal("abc@project.iam.gserviceaccount.com"),
				testutil.WithMode(api.AuditLogRequest_BEST_EFFORT)),
		},
		{
			name: "use_defaults_when_config_file_not_found",
			path: path.Join(dir, "inexistent.yaml"),
			envs: map[string]string{
				"AUDIT_CLIENT_BACKEND_REMOTE_INSECURE_ENABLED":    "true",
				"AUDIT_CLIENT_BACKEND_REMOTE_IMPERSONATE_ACCOUNT": "example@test.iam.gserviceaccount.com",
				"AUDIT_CLIENT_LOG_MODE":                           "LOG_MODE_UNSPECIFIED",
			},
			req: testutil.NewRequest(testutil.WithPrincipal("abc@project.iam.gserviceaccount.com")),
			wantReq: testutil.NewRequest(
				testutil.WithPrincipal("abc@project.iam.gserviceaccount.com"),
				testutil.WithMode(api.AuditLogRequest_FAIL_CLOSE)),
		},
		{
			name:          "invalid_config_file_should_error",
			path:          path.Join(dir, "invalid.yaml"),
			wantErrSubstr: "cannot unmarshal",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := &fakeServer{}

			addr, _ := testutil.TestFakeGRPCServer(t, func(s *grpc.Server) {
				api.RegisterAuditLogAgentServer(s, r)
			})

			l := envconfig.MultiLookuper(
				envconfig.MapLookuper(map[string]string{"AUDIT_CLIENT_BACKEND_REMOTE_ADDRESS": addr}),
				envconfig.MapLookuper(tc.envs),
			)
			c, err := audit.NewClient(ctx, fromConfigFile(tc.path, l))
			if diff := pkgtestutil.DiffErrString(err, tc.wantErrSubstr); diff != "" {
				t.Errorf("audit.NewClient(FromConfigFile(%v)) got unexpected error substring: %v", tc.path, diff)
			}
			if err != nil {
				return
			}
			if err := c.Log(ctx, tc.req); err != nil {
				t.Fatal(err)
			}
			cmpopts := []cmp.Option{
				protocmp.Transform(),
				// We ignore `AuditLog.Metadata` because it contains the
				// runtime information which varies depending on the
				// environment executing the unit test.
				protocmp.IgnoreFields(&capi.AuditLog{}, "metadata"),
			}
			if diff := cmp.Diff(tc.wantReq, r.gotReq, cmpopts...); diff != "" {
				t.Errorf("audit logging backend got request (-want,+got):\n%s", diff)
			}
		})
	}
}

func TestFromConfig(t *testing.T) {
	t.Parallel()

	ctx := logging.WithLogger(t.Context(), logging.TestLogger(t))

	r := &fakeServer{}
	addr, _ := testutil.TestFakeGRPCServer(t, func(s *grpc.Server) {
		api.RegisterAuditLogAgentServer(s, r)
	})

	cases := []struct {
		name          string
		cfg           *api.Config
		req           *api.AuditLogRequest
		wantReq       *api.AuditLogRequest
		wantErrSubstr string
	}{{
		name: "valid_config_success",
		cfg: &api.Config{
			Backend: &api.Backend{
				Remote: &api.Remote{
					InsecureEnabled: true,
					Address:         addr,
				},
			},
		},
		req: testutil.NewRequest(testutil.WithPrincipal("abc@project.iam.gserviceaccount.com")),
		wantReq: testutil.NewRequest(
			testutil.WithPrincipal("abc@project.iam.gserviceaccount.com"),
			testutil.WithMode(api.AuditLogRequest_FAIL_CLOSE)),
	}, {
		name:          "invalid_config_error",
		cfg:           &api.Config{}, // Empty config is invalid
		wantErrSubstr: "invalid configuration:",
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c, err := audit.NewClient(ctx, FromConfig(tc.cfg))
			if diff := pkgtestutil.DiffErrString(err, tc.wantErrSubstr); diff != "" {
				t.Errorf("audit.NewClient(FromConfig(%v)) got unexpected error substring: %v", tc.cfg, diff)
			}
			if err != nil {
				return
			}
			if err := c.Log(ctx, tc.req); err != nil {
				t.Fatal(err)
			}
			cmpopts := []cmp.Option{
				protocmp.Transform(),
				// We ignore `AuditLog.Metadata` because it contains the
				// runtime information which varies depending on the
				// environment executing the unit test.
				protocmp.IgnoreFields(&capi.AuditLog{}, "metadata"),
			}
			if diff := cmp.Diff(tc.wantReq, r.gotReq, cmpopts...); diff != "" {
				t.Errorf("audit logging backend got request (-want,+got):\n%s", diff)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		fileContent string
		wantCfg     *api.Config
	}{
		{
			name: "raw_jwt_with_just_key",
			fileContent: `
version: v1alpha1
backend:
  remote:
    address: foo:443
    insecure_enabled: true
security_context:
  from_raw_jwt:
  - key: authorization
`,
			wantCfg: &api.Config{
				Version:         "v1alpha1",
				LogMode:         api.AuditLogRequest_FAIL_CLOSE.String(),
				Backend:         &api.Backend{Remote: &api.Remote{Address: "foo:443", InsecureEnabled: true}},
				SecurityContext: &api.SecurityContext{FromRawJWT: []*api.FromRawJWT{{Key: "authorization"}}},
			},
		},
		{
			name: "raw_jwt_with_user_defined_values_fully_set",
			fileContent: `
version: v1alpha1
backend:
  remote:
    address: foo:443
    insecure_enabled: true
security_context:
  from_raw_jwt:
  - key: x-jwt-assertion
    prefix: somePrefix
`,
			wantCfg: &api.Config{
				Version:         "v1alpha1",
				LogMode:         api.AuditLogRequest_FAIL_CLOSE.String(),
				Backend:         &api.Backend{Remote: &api.Remote{Address: "foo:443", InsecureEnabled: true}},
				SecurityContext: &api.SecurityContext{FromRawJWT: []*api.FromRawJWT{{Key: "x-jwt-assertion", Prefix: "somePrefix"}}},
			},
		},
		{
			name: "raw_jwt_with_user_defined_values_partially_set",
			fileContent: `
version: v1alpha1
backend:
  remote:
    address: foo:443
    insecure_enabled: true
security_context:
  from_raw_jwt:
  - key: x-jwt-assertion
    prefix:
`,
			wantCfg: &api.Config{
				Version:         "v1alpha1",
				LogMode:         api.AuditLogRequest_FAIL_CLOSE.String(),
				Backend:         &api.Backend{Remote: &api.Remote{Address: "foo:443", InsecureEnabled: true}},
				SecurityContext: &api.SecurityContext{FromRawJWT: []*api.FromRawJWT{{Key: "x-jwt-assertion"}}},
			},
		},
		{
			name: "raw_jwt_with_user_defined_values_partially_set_again",
			fileContent: `
version: v1alpha1
backend:
  remote:
    address: foo:443
    insecure_enabled: true
security_context:
  from_raw_jwt:
  - key: x-jwt-assertion
`,
			wantCfg: &api.Config{
				Version:         "v1alpha1",
				LogMode:         api.AuditLogRequest_FAIL_CLOSE.String(),
				Backend:         &api.Backend{Remote: &api.Remote{Address: "foo:443", InsecureEnabled: true}},
				SecurityContext: &api.SecurityContext{FromRawJWT: []*api.FromRawJWT{{Key: "x-jwt-assertion"}}},
			},
		},
		{
			name: "condition_defined",
			fileContent: `
version: v1alpha1
backend:
  remote:
    address: foo:443
    insecure_enabled: true
condition:
  regex:
    principal_include: "user@example.com"
`,
			wantCfg: &api.Config{
				Version:   "v1alpha1",
				LogMode:   api.AuditLogRequest_FAIL_CLOSE.String(),
				Backend:   &api.Backend{Remote: &api.Remote{Address: "foo:443", InsecureEnabled: true}},
				Condition: &api.Condition{Regex: &api.RegexCondition{PrincipalInclude: "user@example.com"}},
			},
		},
		{
			name: "raw_jwt_with_jvs_and_justification",
			fileContent: `
version: v1alpha1
backend:
  remote:
    address: foo:443
    insecure_enabled: true
security_context:
  from_raw_jwt:
  - key: authorization
justification:
  public_keys_endpoint: example.com:123
  enabled: true
  allow_breakglass: false
`,
			wantCfg: &api.Config{
				Version:         "v1alpha1",
				LogMode:         api.AuditLogRequest_FAIL_CLOSE.String(),
				Backend:         &api.Backend{Remote: &api.Remote{Address: "foo:443", InsecureEnabled: true}},
				SecurityContext: &api.SecurityContext{FromRawJWT: []*api.FromRawJWT{{Key: "authorization"}}},
				Justification: &api.Justification{
					PublicKeysEndpoint: "example.com:123",
					Enabled:            true,
					AllowBreakglass:    false,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "config.yaml")
			if err := os.WriteFile(path, []byte(tc.fileContent), 0o600); err != nil {
				t.Fatal(err)
			}

			fc, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			cfg, err := loadConfig(t.Context(), fc, envconfig.MapLookuper(nil))
			if err != nil {
				t.Errorf("loadConfig() got unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantCfg, cfg); diff != "" {
				t.Errorf("unexpected diff in loadConfig() (-want,+got):\n%s", diff)
			}
		})
	}
}

func TestInterceptorFromConfigFile(t *testing.T) {
	t.Parallel()

	ctx := logging.WithLogger(t.Context(), logging.TestLogger(t))

	cases := []struct {
		name          string
		envs          map[string]string
		fileContent   string
		wantErrSubstr string
	}{
		{
			name: "valid_config_file",
			fileContent: `
version: v1alpha1
backend:
  remote:
    address: foo:443
    insecure_enabled: true
security_context:
  from_raw_jwt:
  - key: "authorization"
    prefix: "Bearer "
rules:
  - selector: "*"
`,
		},
		{
			name: "invalid_config_because_security_context_is_nil",
			// In YAML, empty keys are unset. For details, see:
			// https://stackoverflow.com/a/64462925
			fileContent: `
version: v1alpha1
backend:
  remote:
    address: foo:443
    insecure_enabled: true
security_context:
rules:
  - selector: "*"
`,
			wantErrSubstr: "SecurityContext must be provided to use interceptor",
		},
		{
			name: "invalid_config_due_unset_security_context_again",
			fileContent: `
version: v1alpha1
backend:
  remote:
    address: foo:443
    insecure_enabled: true
rules:
  - selector: "*"
`,
			wantErrSubstr: "SecurityContext must be provided to use interceptor",
		},
		{
			name: "invalid_config_due_missing_jwt_key",
			fileContent: `
version: v1alpha1
backend:
  remote:
    address: foo:443
    insecure_enabled: true
security_context:
  from_raw_jwt:
  - {}
rules:
  - selector: "*"
`,
			wantErrSubstr: "FromRawJWT[0]: key must be specified",
		},
		{
			name: "invalid_config_because_backend_address_is_nil",
			fileContent: `
version: v1alpha1
backend:
  remote:
    address:
    insecure_enabled: true
security_context:
  from_raw_jwt:
  - key: authorization
rules:
  - selector: "*"
`,
			wantErrSubstr: "backend address is nil",
		},
		{
			name: "invalid_config_due_to_log_type",
			fileContent: `
version: v1alpha1
backend:
  remote:
    address: foo:443
    insecure_enabled: true
security_context:
  from_raw_jwt:
  - key: authorization
rules:
  - selector: "*"
    log_type: bananas
`,
			wantErrSubstr: `unexpected rule.LogType "bananas"`,
		},
		{
			name:          "unparsable_config",
			fileContent:   `bananas`,
			wantErrSubstr: "cannot unmarshal",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "config.yaml")
			if err := os.WriteFile(path, []byte(tc.fileContent), 0o600); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
			})

			_, err := audit.NewInterceptor(ctx, interceptorFromConfigFile(path, envconfig.MapLookuper(tc.envs)))
			if diff := pkgtestutil.DiffErrString(err, tc.wantErrSubstr); diff != "" {
				t.Errorf("WithInterceptorFromConfigFile(path) got unexpected error substring: %v", diff)
			}
		})
	}
}
