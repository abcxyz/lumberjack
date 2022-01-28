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
	Version string `yaml:"version,omitempty" env:"VERSION,overwrite"`

	// If a remote backend is omitted, we audit log to stdout.
	Backend *Backend `yaml:"backend,omitempty"`

	// If condition is omitted, the default is to discard logs where
	// the principal is an IAM service account.
	Condition *Condition `yaml:"condition,omitempty"`

	// At the moment, we must require security context.
	SecurityContext *SecurityContext `yaml:"security_context,omitempty"`

	Rules []*AuditRule `yaml:"rules,omitempty"`
}

// Validate checks if the config is valid.
func (cfg *Config) Validate() error {
	// TODO: do validations for each field if necessary.
	var err error
	if cfg.Version != Version {
		err = multierr.Append(err, fmt.Errorf("unexpected Version %q want %q", cfg.Version, Version))
	}
	// TODO(#74): Fall back to stdout logging if backend is nil.
	if cfg.Backend == nil {
		err = multierr.Append(err, fmt.Errorf("backend is nil"))
	} else if serr := cfg.Backend.Validate(); serr != nil {
		err = multierr.Append(err, serr)
	}
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
	if cfg.Condition == nil {
		cfg.Condition = &Condition{}
	}
	cfg.Condition.SetDefault()
	if cfg.SecurityContext != nil {
		cfg.SecurityContext.SetDefault()
	}
	for _, r := range cfg.Rules {
		r.SetDefault()
	}
}

// Backend is the remote backend service to send audit logs.
type Backend struct {
	Address            string `yaml:"address,omitempty" env:"BACKEND_ADDRESS,overwrite"`
	InsecureEnabled    bool   `yaml:"insecure_enabled,omitempty" env:"BACKEND_INSECURE_ENABLED,overwrite"`
	ImpersonateAccount string `yaml:"impersonate_account,omitempty" env:"BACKEND_IMPERSONATE_ACCOUNT,overwrite"`
}

// Validate validates the backend.
func (b *Backend) Validate() error {
	if b.Address == "" {
		return fmt.Errorf("backend address is nil")
	}
	return nil
}

// Condition is the condition to match to collect audit logs.
type Condition struct {
	Regex *RegexCondition `yaml:"regex,omitempty"`
}

// SetDefault sets default for the condition.
func (c *Condition) SetDefault() {
	if c.Regex == nil {
		c.Regex = &RegexCondition{}
	}
	c.Regex.SetDefault()
}

// RegexCondition matches condition with regular expression.
type RegexCondition struct {
	PrincipalExclude string `yaml:"principal_exclude,omitempty" env:"CONDITION_REGEX_PRINCIPAL_EXCLUDE,overwrite"`
	PrincipalInclude string `yaml:"principal_include,omitempty" env:"CONDITION_REGEX_PRINCIPAL_INCLUDE,overwrite"`
}

// SetDefault sets default for the regex condition.
func (rc *RegexCondition) SetDefault() {
	// By default, we exclude any service accounts from audit logging.
	if rc.PrincipalInclude == "" && rc.PrincipalExclude == "" {
		rc.PrincipalExclude = ".gserviceaccount.com$"
	}
}

// SecurityContext provides instructive info for where to retrieve
// the security context, e.g. authentication info.
type SecurityContext struct {
	FromRawJWT *FromRawJWT `yaml:"from_raw_jwt,omitempty"`
}

// Validate validates the security context.
func (sc *SecurityContext) Validate() error {
	if sc.FromRawJWT == nil {
		return fmt.Errorf("one and only one SecurityContext option must be specified")
	}
	return nil
}

// SetDefault sets default for the security context.
func (sc *SecurityContext) SetDefault() {
	if sc.FromRawJWT != nil {
		sc.FromRawJWT.SetDefault()
	}
}

// FromRawJWT provides info for how to retrieve security context from
// a raw JWT.
type FromRawJWT struct {
	Key    string `yaml:"key,omitempty"`
	Prefix string `yaml:"prefix,omitempty"`
	JWKs   *JWKs  `yaml:"jwks,omitempty"`
}

// SetDefault sets default for the JWT security context.
func (j *FromRawJWT) SetDefault() {
	if j.Key == "" && j.Prefix == "" {
		j.Key = "authorization"
		j.Prefix = "Bearer "
	}
}

// JWKs provides JWKs to validate a JWT.
type JWKs struct {
	Endpoint string `yaml:"endpoint,omitempty"`
}

// AuditRule is an audit rule to instruct how to audit selected paths/methods.
type AuditRule struct {
	Selector  string `yaml:"selector,omitempty"`
	Directive string `yaml:"directive,omitempty"`
	LogType   string `yaml:"log_type,omitempty"`
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
