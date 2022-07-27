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
	"fmt"

	"go.uber.org/multierr"
)

const (
	// Version of the API and config.
	Version = "v1alpha1"

	// Audit rule directive options.
	AuditRuleDirectiveDefault            = "AUDIT"
	AuditRuleDirectiveRequestOnly        = "AUDIT_REQUEST_ONLY"
	AuditRuleDirectiveRequestAndResponse = "AUDIT_REQUEST_AND_RESPONSE"
)

// Config is the full audit client config.
type Config struct {
	// Version is the version of the config.
	Version string `yaml:"version,omitempty" env:"VERSION,overwrite"`

	// Backend specifies what remote backend to send audit logs to.
	// If a remote backend config is nil, audit logs will be written to stdout.
	Backend *Backend `yaml:"backend,omitempty" env:",noinit"`

	// Condition specifies the condition under which an incoming request should be
	// audit logged. If the condition is nil, the default is to audit log all requests.
	Condition *Condition `yaml:"condition,omitempty" env:",noinit"`

	// SecurityContext specifies how to retrieve security context such as
	// authentication info from the incoming requests.
	// This config is only used for auto audit logging, and it must not be nil.
	// When auto audit logging is not used, setting this field has no effect.
	SecurityContext *SecurityContext `yaml:"security_context,omitempty" env:",noinit"`

	// Rules specifies audit logging instructions per matching requests
	// method/path. If the rules is nil or empty, no audit logs will be collected.
	// This config is only used for auto audit logging.
	// When auto audit logging is not used, setting this field has no effect.
	Rules []*AuditRule `yaml:"rules,omitempty"`

	// Labels are additional labels that the calling code wants added to each
	// audit log request. Each label will only be added if it is not already added
	// in the audit log, and will not overwrite explicitly added labels.
	Labels map[string]string `yaml:"labels,omitempty"`

	// LogMode specifies whether the audit logger should fail open or close.
	// If fail-close is not chosen, the audit logger will log errors that occur,
	// and then continue without impeding the application in any way.
	LogMode string `yaml:"log_mode,omitempty" env:"LOG_MODE,overwrite"`

	// JVSEndpoint is the endpoint where public keys may be retrieved from the JVS.
	// These will be used to validate JWT tokens that are passed in through the
	// "justification-token" header.
	JVSEndpoint string `yaml:"jvs_endpoint,omitempty" env:"JVS_ENDPOINT,overwrite,default=localhost:8080"`

	// RequireJustification enables adding justification information to audit logs. If this is enabled,
	// all manual calls are expected to pass in a justification in the "justification-token" header.
	// If omitted, justifications will not be added to logs, even if provided.
	RequireJustification bool `yaml:"require_justification,omitempty" env:"REQUIRE_JUSTIFICATION,overwrite,default=false"`
}

// Validate checks if the config is valid.
func (cfg *Config) Validate() error {
	// TODO: do validations for each field if necessary.
	var err error
	if cfg.Version != Version {
		err = multierr.Append(err, fmt.Errorf("unexpected Version %q want %q", cfg.Version, Version))
	}

	if cfg.Backend == nil {
		err = multierr.Append(err, fmt.Errorf("backend is nil"))
	} else if serr := cfg.Backend.Validate(); serr != nil {
		err = multierr.Append(err, serr)
	}

	if cfg.SecurityContext != nil {
		if serr := cfg.SecurityContext.Validate(); serr != nil {
			err = multierr.Append(err, serr)
		}
	}

	for _, r := range cfg.Rules {
		if rerr := r.Validate(); rerr != nil {
			err = multierr.Append(err, rerr)
		}
	}

	if cfg.LogMode != "" {
		if _, ok := AuditLogRequest_LogMode_value[cfg.LogMode]; !ok {
			err = multierr.Append(err, fmt.Errorf("invalid LogMode %q", cfg.LogMode))
		}
	}

	return err
}

// ValidateSecurityContext checks if the config SecurityContext is valid.
func (cfg *Config) ValidateSecurityContext() error {
	if cfg.SecurityContext == nil {
		return fmt.Errorf("SecurityContext is nil")
	}
	return cfg.SecurityContext.Validate()
}

// SetDefault sets default for the config.
func (cfg *Config) SetDefault() {
	// TODO: set defaults for other fields if necessary.
	if cfg.Version == "" {
		cfg.Version = Version
	}

	// TODO(#74): set default backend to logging to stdout.
	if cfg.Backend != nil {
		cfg.Backend.SetDefault()
	}

	// TODO: set defaults for SecurityContext and Condition
	// once we have any such logic.
	for _, r := range cfg.Rules {
		r.SetDefault()
	}
}

// GetLogMode converts the LogMode string to a AuditLogRequest_LogMode.
func (cfg *Config) GetLogMode() AuditLogRequest_LogMode {
	return AuditLogRequest_LogMode(AuditLogRequest_LogMode_value[cfg.LogMode])
}

// Backend holds information on the backends to send logs to.
type Backend struct {
	Remote       *Remote       `yaml:"remote,omitempty" env:",noinit"`
	CloudLogging *CloudLogging `yaml:"cloudlogging,omitempty" env:",noinit"`
}

// SetDefault sets default for the Backend.
func (b *Backend) SetDefault() {
	if b.CloudLogging != nil {
		b.CloudLogging.SetDefault()
	}
}

// Validate validates the Backend.
func (b *Backend) Validate() error {
	backendSet := false
	var merr error
	if b.Remote != nil {
		backendSet = true
		if err := b.Remote.Validate(); err != nil {
			merr = multierr.Append(merr, err)
		}
	}
	if b.CloudLogging != nil {
		backendSet = true
		if err := b.CloudLogging.Validate(); err != nil {
			merr = multierr.Append(merr, err)
		}
	}
	if !backendSet {
		merr = multierr.Append(merr, fmt.Errorf("no backend is set"))
	}
	return merr
}

// Remote is the remote backend service to send audit logs to.
// The backend must be a gRPC service that implements protos/v1alpha1/audit_log_agent.proto.
type Remote struct {
	// Address is the remote backend address. It must be set.
	Address string `yaml:"address,omitempty" env:"BACKEND_REMOTE_ADDRESS,overwrite"`

	// InsecureEnabled indicates whether to insecurely connect to the backend.
	// This should be set to false for production usage.
	InsecureEnabled bool `yaml:"insecure_enabled,omitempty" env:"BACKEND_REMOTE_INSECURE_ENABLED,overwrite"`

	// ImpersonateAccount specifies which service account to impersonate to call the backend.
	// If empty, there will be no impersonation.
	ImpersonateAccount string `yaml:"impersonate_account,omitempty" env:"BACKEND_REMOTE_IMPERSONATE_ACCOUNT,overwrite"`
}

// Validate validates the backend.
func (b *Remote) Validate() error {
	if b.Address == "" {
		return fmt.Errorf("backend address is nil")
	}
	return nil
}

// CloudLogging is the GCP cloud logging backend to send audit logs to.
type CloudLogging struct {
	// DefaultProject indicates whether to use the project where the client runs.
	DefaultProject bool `yaml:"default_project,omitempty" env:"BACKEND_CLOUDLOGGING_DEFAULT_PROJECT,overwrite"`

	// Project allows overriding the project where to send the audit logs.
	// The client must be run with a service account that has log writer role on the project.
	Project string `yaml:"project,omitempty" env:"BACKEND_CLOUDLOGGING_PROJECT,overwrite"`
}

// SetDefault sets default on the CloudLogging backend.
func (cl *CloudLogging) SetDefault() {
	if cl.Project == "" && !cl.DefaultProject {
		cl.DefaultProject = true
	}
}

// Validate validates the CloudLogging backend.
func (cl *CloudLogging) Validate() error {
	if cl.Project == "" && !cl.DefaultProject {
		return fmt.Errorf("backend cloudlogging no project or using default project is set")
	}
	if cl.Project != "" && cl.DefaultProject {
		return fmt.Errorf("backend cloudlogging project is set while using default project")
	}
	return nil
}

// Condition is the condition the condition under which an incoming request should be
// audit logged. Only one condition can be used.
type Condition struct {
	// Regex specifies the regular experessions to match request principals.
	Regex *RegexCondition `yaml:"regex,omitempty" env:",noinit"`
}

// RegexCondition matches condition with regular expression.
// If PrincipalInclude and PrincipalExclude are both empty, all requests will be audit logged.
// When only PrincipalInclude is set, only the matching requests will be audit logged.
// When only PrincipalExclude is set, only the non-matching requests will be audit logged.
// When both PrincipalInclude and PrincipalExclude are both set, PrincipalInclude takes precedence.
// If a request matches PrincipalInclude, it will be audit logged no matter whether it matches
// PrincipalExclude.
type RegexCondition struct {
	// PrincipalInclude specifies a regular expression to match request principals to be included in audit logging.
	PrincipalInclude string `yaml:"principal_include,omitempty" env:"CONDITION_REGEX_PRINCIPAL_INCLUDE,overwrite"`
	// PrincipalExclude specifies a regular expression to match request principals to be excluded from audit logging.
	PrincipalExclude string `yaml:"principal_exclude,omitempty" env:"CONDITION_REGEX_PRINCIPAL_EXCLUDE,overwrite"`
}

// SecurityContext provides instructive info for where to retrieve
// the security context, e.g. authentication info.
type SecurityContext struct {
	// FromRawJWT specifies where to look up the JWT.
	FromRawJWT []*FromRawJWT `yaml:"from_raw_jwt,omitempty"`
}

// Validate validates the security context.
func (sc *SecurityContext) Validate() error {
	if len(sc.FromRawJWT) == 0 {
		return fmt.Errorf("one and only one SecurityContext option must be specified")
	}

	var merr error
	for i, j := range sc.FromRawJWT {
		if err := j.Validate(); err != nil {
			merr = multierr.Append(merr, fmt.Errorf("FromRawJWT[%d]: %w", i, err))
		}
	}
	return merr
}

// FromRawJWT provides info for how to retrieve security context from
// a raw JWT.
type FromRawJWT struct {
	// Key is the metadata key whose value is a JWT.
	Key string `yaml:"key,omitempty"`
	// Prefix is the prefix to truncate the metadata value
	// to retrieve the JWT.
	Prefix string `yaml:"prefix,omitempty"`
	// JWKs specifies the JWKs to validate the JWT.
	// If JWTs is nil, the JWT won't be validated.
	JWKs *JWKs `yaml:"jwks,omitempty"`
}

// Validate validates the FromRawJWT.
func (j *FromRawJWT) Validate() error {
	if j.Key == "" {
		return fmt.Errorf("key must be specified")
	}
	return nil
}

// JWKs provides JWKs to validate a JWT.
type JWKs struct {
	// Endpoint is the endpoint to retrieve the JWKs to validate JWT.
	Endpoint string `yaml:"endpoint,omitempty"`
}

// AuditRule is an audit rule to instruct how to audit selected paths/methods.
type AuditRule struct {
	// Selector is a string to match request methods/paths.
	// In gRPC, this is in the format of "/[service_name].[method_name]".
	Selector string `yaml:"selector,omitempty"`

	// Directive specifies what audit action to take for the matching requests.
	// Allowed values are:
	// "AUDIT" - write audit log without request/response.
	// "AUDIT_REQUEST_ONLY" - write audit log with only request.
	// "AUDIT_REQUEST_AND_RESPONSE" - write audit log with request and response.
	Directive string `yaml:"directive,omitempty"`

	// LogType specifies the audit log type for the matching requests.
	// Allowed values are:
	// "ADMIN_ACTIVITY" - the access is an admin operation
	// "DATA_ACCESS" - the access is a data access
	// If empty, the default value is "DATA_ACCESS".
	LogType string `yaml:"log_type,omitempty"`
}

// Validate validates the audit rule.
func (r *AuditRule) Validate() error {
	if r.Selector == "" {
		return fmt.Errorf("audit rule selector is empty")
	}
	if r.Directive == "" {
		return fmt.Errorf("audit rule directive is empty")
	}
	if r.LogType == "" {
		return fmt.Errorf("audit rule log type is empty")
	}
	switch r.Directive {
	case AuditRuleDirectiveDefault:
	case AuditRuleDirectiveRequestOnly:
	case AuditRuleDirectiveRequestAndResponse:
	default:
		return fmt.Errorf("unexpected rule.Directive %q want one of [%q, %q, %q]",
			r.Directive, AuditRuleDirectiveDefault, AuditRuleDirectiveRequestOnly, AuditRuleDirectiveRequestAndResponse)
	}
	switch r.LogType {
	case AuditLogRequest_ADMIN_ACTIVITY.String():
	case AuditLogRequest_DATA_ACCESS.String():
	default:
		return fmt.Errorf("unexpected rule.LogType %q want one of [%q, %q]",
			r.LogType, AuditLogRequest_ADMIN_ACTIVITY.String(), AuditLogRequest_DATA_ACCESS.String())
	}
	return nil
}

// SetDefault sets default for the audit rule.
func (r *AuditRule) SetDefault() {
	if r.Directive == "" {
		r.Directive = AuditRuleDirectiveDefault
	}
	if r.LogType == "" {
		r.LogType = AuditLogRequest_DATA_ACCESS.String()
	}
}
