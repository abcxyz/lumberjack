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
	Backend *Backend `yaml:"backend,omitempty"`

	// Condition specifies the condition under which an incoming request should be
	// audit logged. If the condition is nil, the default is to audit log all requests.
	Condition *Condition `yaml:"condition,omitempty"`

	// SecurityContext specifies how to retrieve security context such as
	// authentication info from the incoming requests.
	// This config is only used for auto audit logging, and it must not be nil.
	// When auto audit logging is not used, setting this field has no effect.
	SecurityContext *SecurityContext `yaml:"security_context,omitempty"`

	// Rules specifies audit logging instructions per matching requests
	// method/path. If the rules is nil or empty, no audit logs will be collected.
	// This config is only used for auto audit logging.
	// When auto audit logging is not used, setting this field has no effect.
	Rules []*AuditRule `yaml:"rules,omitempty"`

	// LogMode specifies whether the audit logger should fail open or close.
	// If fail-close is not chosen, the audit logger will try to swallow any
	// errors that occur, and not impede the application in any way.
	LogMode AuditLogRequest_LogMode `yaml: "log_mode,omitempty"`
}

// Validate checks if the config is valid.
func (cfg *Config) Validate() error {
	// TODO: do validations for each field if necessary.
	var err error
	if cfg.Version != Version {
		err = multierr.Append(err, fmt.Errorf("unexpected Version %q want %q", cfg.Version, Version))
	}

	if cfg.Backend == nil {
		// TODO(#74): Fall back to stdout logging if backend is nil.
		err = multierr.Append(err, fmt.Errorf("backend is nil"))
	} else if serr := cfg.Backend.Validate(); serr != nil {
		err = multierr.Append(err, serr)
	}

	// TODO(#123): Reenable this validation once we stop initiate nil structs.
	// if cfg.SecurityContext != nil {
	// 	if serr := cfg.SecurityContext.Validate(); serr != nil {
	// 		err = multierr.Append(err, serr)
	// 	}
	// }

	for _, r := range cfg.Rules {
		if rerr := r.Validate(); rerr != nil {
			err = multierr.Append(err, rerr)
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
	// TODO: set defaults for SecurityContext and Condition
	// once we have any such logic.
	for _, r := range cfg.Rules {
		r.SetDefault()
	}
}

// ShouldFailClose returns true only if FAIL_CLOSE is explicitly configured. On BEST_EFFORT or LOG_MODE_UNSPECIFIED (the default) then return false.
func (cfg *Config) ShouldFailClose() bool {
	return cfg.LogMode == AuditLogRequest_FAIL_CLOSE
}

// Backend is the remote backend service to send audit logs to.
// The backend must be a gRPC service that implements protos/v1alpha1/audit_log_agent.proto.
type Backend struct {
	// Address is the remote backend address. It must be set.
	Address string `yaml:"address,omitempty" env:"BACKEND_ADDRESS,overwrite"`

	// InsecureEnabled indicates whether to insecurely connect to the backend.
	// This should be set to false for production usage.
	InsecureEnabled bool `yaml:"insecure_enabled,omitempty" env:"BACKEND_INSECURE_ENABLED,overwrite"`

	// ImpersonateAccount specifies which service account to impersonate to call the backend.
	// If empty, there will be no impersonation.
	ImpersonateAccount string `yaml:"impersonate_account,omitempty" env:"BACKEND_IMPERSONATE_ACCOUNT,overwrite"`
}

// Validate validates the backend.
func (b *Backend) Validate() error {
	if b.Address == "" {
		return fmt.Errorf("backend address is nil")
	}
	return nil
}

// Condition is the condition the condition under which an incoming request should be
// audit logged. Only one condition can be used.
type Condition struct {
	// Regex specifies the regular experessions to match request principals.
	Regex *RegexCondition `yaml:"regex,omitempty"`
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
			merr = multierr.Append(merr, fmt.Errorf("FromRawJWT[%d]: %v", i, err))
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

// Validate validates the FromRawJWT
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
