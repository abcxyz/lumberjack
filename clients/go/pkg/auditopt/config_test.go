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
			go s.Serve(lis)

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
			name: "include_and_exclude_in_filter",
			path: path.Join(dir, "inexistent2.yaml"),
			envs: map[string]string{
				"AUDIT_CLIENT_FILTER_REGEX_PRINCIPAL_INCLUDE": "iam.gserviceaccount.com$",
				"AUDIT_CLIENT_FILTER_REGEX_PRINCIPAL_EXCLUDE": "iam.gserviceaccount.com$",
				"AUDIT_CLIENT_BACKEND_INSECURE_ENABLED":       "true",
			},
			req:     testutil.ReqBuilder().WithPrincipal("abc@project.iam.gserviceaccount.com").Build(),
			wantReq: testutil.ReqBuilder().WithPrincipal("abc@project.iam.gserviceaccount.com").Build(),
		},
		{
			name: "include_overwrites_default_exclude",
			path: path.Join(dir, "inexistent2.yaml"),
			envs: map[string]string{
				"AUDIT_CLIENT_FILTER_REGEX_PRINCIPAL_INCLUDE": "iam.gserviceaccount.com$",
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
			go s.Serve(lis)

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
