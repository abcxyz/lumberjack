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
	"fmt"
	"io/ioutil"
	"net"
	"path"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	calpb "google.golang.org/genproto/googleapis/cloud/audit"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/testing/protocmp"

	alpb "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
	"github.com/abcxyz/lumberjack/clients/go/pkg/audit"
	"github.com/abcxyz/lumberjack/clients/go/pkg/errutil"
	"github.com/abcxyz/lumberjack/clients/go/pkg/security"
	"github.com/abcxyz/lumberjack/clients/go/pkg/testutil"
)

type fakeServer struct {
	alpb.UnimplementedAuditLogAgentServer
	gotReq *alpb.AuditLogRequest
}

func (s *fakeServer) ProcessLog(_ context.Context, logReq *alpb.AuditLogRequest) (*alpb.AuditLogResponse, error) {
	s.gotReq = logReq
	return &alpb.AuditLogResponse{Result: logReq}, nil
}

func TestMustFromConfigFile(t *testing.T) {
	// No parallel since testing with env vars.
	cases := []struct {
		name          string
		envs          map[string]string
		fileContent   string
		req           *alpb.AuditLogRequest
		wantReq       *alpb.AuditLogRequest
		wantErrSubstr string
	}{
		{
			name: "use_default_when_principal_exclude_unset",
			// In YAML, empty keys are unset. For details, see:
			// https://stackoverflow.com/a/64462925
			fileContent: `
version: v1alpha1
filter:
  regex:
    principal_exclude: # unset
backend:
  address: %s
  insecure_enabled: true
`,
			// By default, we ignore log requests that have an IAM service account
			// as a principal.
			req: testutil.ReqBuilder().WithPrincipal("abc@project.iam.gserviceaccount.com").Build(),
		},
		{
			name: "overwrite_default_when_principal_exclude_set",
			fileContent: `
version: v1alpha1
filter:
  regex:
    principal_exclude: user@example.com$
backend:
  address: %s
  insecure_enabled: true
`,
			req:     testutil.ReqBuilder().WithPrincipal("abc@project.iam.gserviceaccount.com").Build(),
			wantReq: testutil.ReqBuilder().WithPrincipal("abc@project.iam.gserviceaccount.com").Build(),
		},
		{
			name: "env_var_overwrites_config_file",
			envs: map[string]string{
				"AUDIT_CLIENT_FILTER_REGEX_PRINCIPAL_EXCLUDE": "user@example.com",
			},
			fileContent: `
version: v1alpha1
filter:
  regex:
    principal_exclude: abc@project.iam.gserviceaccount.com$
backend:
  address: %s
  insecure_enabled: true
`,
			req:     testutil.ReqBuilder().WithPrincipal("abc@project.iam.gserviceaccount.com").Build(),
			wantReq: testutil.ReqBuilder().WithPrincipal("abc@project.iam.gserviceaccount.com").Build(),
		},
		{
			name: "explicitly_include_service_account",
			fileContent: `
version: v1alpha1
filter:
  regex:
    principal_include: abc@project.iam.gserviceaccount.com
    principal_exclude: .iam.gserviceaccount.com$
backend:
  address: %s
  insecure_enabled: true
`,
			req:     testutil.ReqBuilder().WithPrincipal("abc@project.iam.gserviceaccount.com").Build(),
			wantReq: testutil.ReqBuilder().WithPrincipal("abc@project.iam.gserviceaccount.com").Build(),
		},
		{
			name: "empty_string_allowed_for_regex_filter",
			fileContent: `
version: v1alpha1
filter:
  regex:
    principal_include: ""
    principal_exclude: .iam.gserviceaccount.com$
backend:
  address: %s
  insecure_enabled: true
`,
			req: testutil.ReqBuilder().WithPrincipal("abc@project.iam.gserviceaccount.com").Build(),
		},
		{
			name:          "invalid_config_file_should_error",
			fileContent:   `bananas`,
			wantErrSubstr: "cannot unmarshal",
		},
		{
			name: "nil_backend_address_should_error",
			envs: map[string]string{
				"AUDIT_CLIENT_BACKEND_ADDRESS": "",
			},
			fileContent: `
version: v1alpha1
noop: %s
`,
			wantErrSubstr: "config backend address is nil, set it as an env var or in a config file",
		},
		{
			name: "wrong_version_should_error",
			fileContent: `
version: v2
backend:
  address: %s
`,
			wantErrSubstr: `config version "v2" unsupported, supported versions are ["v1alpha1"]`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := &fakeServer{}
			s := grpc.NewServer()
			defer s.Stop()
			alpb.RegisterAuditLogAgentServer(s, r)

			lis, err := net.Listen("tcp", "localhost:0")
			if err != nil {
				t.Fatalf("net.Listen(tcp, localhost:0) failed: %v", err)
			}
			go func(t *testing.T, s *grpc.Server, lis net.Listener) {
				err := s.Serve(lis)
				if err != nil {
					// TODO: see about using Errorf here instead of Logf
					// doing a logf here creates a data race under intergration testing.
					fmt.Printf("net.Listen(tcp, localhost:0) serve failed: %v", err)
					// t.Logf("net.Listen(tcp, localhost:0) serve failed: %v", err)
				}
			}(t, s, lis)

			for k, v := range tc.envs {
				t.Setenv(k, v)
			}

			path := filepath.Join(t.TempDir(), "config.yaml")
			// Add address of the fake server to the config file.
			content := fmt.Sprintf(tc.fileContent, lis.Addr().String())
			if err := ioutil.WriteFile(path, []byte(content), 0o600); err != nil {
				t.Fatalf("error creating test config file: %v", err)
			}

			c, err := audit.NewClient(MustFromConfigFile(path))
			if diff := errutil.DiffSubstring(err, tc.wantErrSubstr); diff != "" {
				t.Errorf("audit.NewClient(FromConfigFile(%v)) got unexpected error substring: %v", path, diff)
			}
			if err != nil {
				return
			}
			if err := c.Log(context.Background(), tc.req); err != nil {
				t.Fatalf("client.Log(...) unexpected error: %v", err)
			}
			cmpopts := []cmp.Option{
				protocmp.Transform(),
				// We ignore `AuditLog.Metadata` because it contains the
				// runtime information which varies depending on the
				// environment executing the unit test.
				protocmp.IgnoreFields(&calpb.AuditLog{}, "metadata"),
			}
			if diff := cmp.Diff(tc.wantReq, r.gotReq, cmpopts...); diff != "" {
				t.Errorf("audit logging backend got request (-want,+got):\n%s", diff)
			}
		})
	}
}

func TestFromConfigFile(t *testing.T) {
	// No parallel since testing with env vars.
	configFileContentByName := map[string]string{
		"valid.yaml": `
version: v1alpha1
filter:
  regex:
    principal_exclude: user@example.com$
backend:
  insecure_enabled: true
`,
		"invalid.yaml": `bananas`,
	}
	// Set up config files.
	dir := t.TempDir()
	for name, content := range configFileContentByName {
		path := filepath.Join(dir, name)
		if err := ioutil.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("error creating test config file: %v", err)
		}
	}
	cases := []struct {
		name          string
		envs          map[string]string
		path          string
		req           *alpb.AuditLogRequest
		wantReq       *alpb.AuditLogRequest
		wantErrSubstr string
	}{
		{
			name:    "use_config_file_when_provided",
			path:    path.Join(dir, "valid.yaml"),
			req:     testutil.ReqBuilder().WithPrincipal("abc@project.iam.gserviceaccount.com").Build(),
			wantReq: testutil.ReqBuilder().WithPrincipal("abc@project.iam.gserviceaccount.com").Build(),
		},
		{
			name: "use_env_var_when_config_file_not_found",
			path: path.Join(dir, "inexistent.yaml"),
			envs: map[string]string{
				"AUDIT_CLIENT_FILTER_REGEX_PRINCIPAL_EXCLUDE": "user@example.com$",
				"AUDIT_CLIENT_BACKEND_INSECURE_ENABLED":       "true",
			},
			req:     testutil.ReqBuilder().WithPrincipal("abc@project.iam.gserviceaccount.com").Build(),
			wantReq: testutil.ReqBuilder().WithPrincipal("abc@project.iam.gserviceaccount.com").Build(),
		},
		{
			name:          "invalid_config_file_should_error",
			path:          path.Join(dir, "invalid.yaml"),
			wantErrSubstr: "cannot unmarshal",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := &fakeServer{}
			s := grpc.NewServer()
			defer s.Stop()
			alpb.RegisterAuditLogAgentServer(s, r)

			lis, err := net.Listen("tcp", "localhost:0")
			if err != nil {
				t.Fatalf("net.Listen(tcp, localhost:0) failed: %v", err)
			}
			go func(t *testing.T, s *grpc.Server, lis net.Listener) {
				err := s.Serve(lis)
				if err != nil {
					// TODO: see about using Errorf here instead of Logf
					t.Logf("net.Listen(tcp, localhost:0) serve failed: %v", err)
				}
			}(t, s, lis)

			t.Setenv("AUDIT_CLIENT_BACKEND_ADDRESS", lis.Addr().String())
			for k, v := range tc.envs {
				t.Setenv(k, v)
			}

			c, err := audit.NewClient(FromConfigFile(tc.path))
			if diff := errutil.DiffSubstring(err, tc.wantErrSubstr); diff != "" {
				t.Errorf("audit.NewClient(FromConfigFile(%v)) got unexpected error substring: %v", tc.path, diff)
			}
			if err != nil {
				return
			}
			if err := c.Log(context.Background(), tc.req); err != nil {
				t.Fatalf("client.Log(...) unexpected error: %v", err)
			}
			cmpopts := []cmp.Option{
				protocmp.Transform(),
				// We ignore `AuditLog.Metadata` because it contains the
				// runtime information which varies depending on the
				// environment executing the unit test.
				protocmp.IgnoreFields(&calpb.AuditLog{}, "metadata"),
			}
			if diff := cmp.Diff(tc.wantReq, r.gotReq, cmpopts...); diff != "" {
				t.Errorf("audit logging backend got request (-want,+got):\n%s", diff)
			}
		})
	}
}

func TestFromRawJWTFromViper(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name           string
		fileContent    string
		wantFromRawJWT *security.FromRawJWT
		wantErrSubstr  string
	}{
		{
			name: "unset_security_context",
			fileContent: `
version: v1alpha1
backend:
  address: foo:443
  insecure_enabled: true
security_context:
`,
		},
		{
			name: "unset_security_context_again",
			fileContent: `
version: v1alpha1
backend:
  address: foo:443
  insecure_enabled: true
`,
		},
		{
			name: "raw_jwt_with_default_value_due_to_braces",
			fileContent: `
version: v1alpha1
backend:
  address: foo:443
  insecure_enabled: true
security_context:
  from_raw_jwt: {}
`,
			wantFromRawJWT: &security.FromRawJWT{
				Key:    "authorization",
				Prefix: "bearer ",
			},
		},
		{
			name: "raw_jwt_with_default_value_due_to_null",
			fileContent: `
version: v1alpha1
backend:
  address: foo:443
  insecure_enabled: true
security_context:
  from_raw_jwt:
`,
			wantFromRawJWT: &security.FromRawJWT{
				Key:    "authorization",
				Prefix: "bearer ",
			},
		},
		{
			name: "raw_jwt_with_default_value_due_to_empty_string",
			fileContent: `
version: v1alpha1
backend:
  address: foo:443
  insecure_enabled: true
security_context:
  from_raw_jwt:
    key: ""
    prefix: ""
`,
			wantFromRawJWT: &security.FromRawJWT{
				Key:    "authorization",
				Prefix: "bearer ",
			},
		},
		{
			name: "raw_jwt_with_user-defined_values_fully_set",
			fileContent: `
version: v1alpha1
backend:
  address: foo:443
  insecure_enabled: true
security_context:
  from_raw_jwt:
    key: x-jwt-assertion
    prefix: somePrefix
`,
			wantFromRawJWT: &security.FromRawJWT{
				Key:    "x-jwt-assertion",
				Prefix: "somePrefix",
			},
		},
		{
			name: "raw_jwt_with_user-defined_values_partially_set",
			fileContent: `
version: v1alpha1
backend:
  address: foo:443
  insecure_enabled: true
security_context:
  from_raw_jwt:
    key: x-jwt-assertion
    prefix:
`,
			wantFromRawJWT: &security.FromRawJWT{
				Key:    "x-jwt-assertion",
				Prefix: "",
			},
		},
		{
			name: "raw_jwt_with_user-defined_values_partially_set_again",
			fileContent: `
version: v1alpha1
backend:
  address: foo:443
  insecure_enabled: true
security_context:
  from_raw_jwt:
    key: x-jwt-assertion
`,
			wantFromRawJWT: &security.FromRawJWT{
				Key:    "x-jwt-assertion",
				Prefix: "",
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "config.yaml")
			if err := ioutil.WriteFile(path, []byte(tc.fileContent), 0o600); err != nil {
				t.Fatalf("error creating test config file: %v", err)
			}
			v := prepareViper()
			if err := setupViperConfigFile(v, path); err != nil {
				t.Fatal(err)
			}

			fromRawJWT, err := fromRawJWTFromViper(v)
			if diff := errutil.DiffSubstring(err, tc.wantErrSubstr); diff != "" {
				t.Errorf("fromRawJWTFromViper(path) got unexpected error substring: %v", diff)
			}
			if diff := cmp.Diff(tc.wantFromRawJWT, fromRawJWT); diff != "" {
				t.Errorf("unexpected diff in fromRawJWT (-want,+got):\n%s", diff)
			}
		})
	}
}

func TestWithInterceptorFromConfigFile(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name          string
		fileContent   string
		wantErrSubstr string
	}{
		{
			name: "valid_config_file",
			fileContent: `
version: v1alpha1
backend:
  address: foo:443
  insecure_enabled: true
security_context:
  from_raw_jwt: {}
`,
		},
		{
			name: "invalid_config_because_security_context_is_nil",
			// In YAML, empty keys are unset. For details, see:
			// https://stackoverflow.com/a/64462925
			fileContent: `
version: v1alpha1
backend:
  address: foo:443
  insecure_enabled: true
security_context:
`,
			wantErrSubstr: `no supported security context configured in config file`,
		},
		{
			name: "invalid_config_because_backend_address_is_nil",
			fileContent: `
version: v1alpha1
backend:
  address:
  insecure_enabled: true
security_context:
  from_raw_jwt: {}
`,
			wantErrSubstr: "failed to create audit client from config file",
		},
		{
			name:          "unparsable_config",
			fileContent:   `bananas`,
			wantErrSubstr: `failed to setup viper from config file`,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "config.yaml")
			if err := ioutil.WriteFile(path, []byte(tc.fileContent), 0o600); err != nil {
				t.Fatalf("error creating test config file: %v", err)
			}

			_, _, err := WithInterceptorFromConfigFile(path)
			if diff := errutil.DiffSubstring(err, tc.wantErrSubstr); diff != "" {
				t.Errorf("WithInterceptorFromConfigFile(path) got unexpected error substring: %v", diff)
			}
		})
	}
}
