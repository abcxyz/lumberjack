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
	Version string `mapstructure:"version,omitempty" json:"version,omitempty"`

	// If a remote backend is omitted, we audit log to stdout.
	Backend *Backend `mapstructure:"backend,omitempty" json:"backend,omitempty"`

	// If condition is omitted, the default is to discard logs where
	// the principal is an IAM service account.
	Condition *Condition `mapstructure:"condition,omitempty" json:"condition,omitempty"`

	// At the moment, we must require security context.
	SecurityContext *SecurityContext `mapstructure:"security_context,omitempty" json:"security_context,omitempty"`

	Rules []*AuditRule `mapstructure:"rules,omitempty" json:"rules,omitempty"`
}

// Validate checks if the config is valid.
func (cfg *Config) Validate() error {
	// TODO: do validations for each field if necessary.
	var err error
	if cfg.Version != Version {
		err = multierr.Append(err, fmt.Errorf("unexpected Version %q", cfg.Version))
	}
	if cfg.SecurityContext == nil {
		err = multierr.Append(err, fmt.Errorf("SecurityContext is nil"))
	} else if serr := cfg.SecurityContext.Validate(); serr != nil {
		err = multierr.Append(err, serr)
	}
	for _, r := range cfg.Rules {
		if rerr := r.Validate(); rerr != nil {
			err = multierr.Append(err, rerr)
		}
	}
	return err
}

// SetDefault sets default for the config.
func (cfg *Config) SetDefault() {
	// TODO: set defaults for other fields if necessary.
	if cfg.Version == "" {
		cfg.Version = Version
	}
	if cfg.SecurityContext != nil {
		cfg.SecurityContext.SetDefault()
	}
	if cfg.Condition != nil {
		cfg.Condition.SetDefault()
	}
	for _, r := range cfg.Rules {
		r.SetDefault()
	}
}

// Backend is the remote backend service to send audit logs.
type Backend struct {
	Address            string `mapstructure:"address,omitempty" json:"address,omitempty"`
	InsecureEnabled    bool   `mapstructure:"insecure_enabled,omitempty" json:"insecure_enabled,omitempty"`
	ImpersonateAccount string `mapstructure:"impersonate_account,omitempty" json:"impersonate_account,omitempty"`
}

// Condition is the condition to match to collect audit logs.
type Condition struct {
	Regex *RegexCondition `mapstructure:"regex,omitempty" json:"regex,omitempty"`
}

// SetDefault sets default for the condition.
func (c *Condition) SetDefault() {
	if c.Regex != nil {
		c.Regex.SetDefault()
	}
}

// RegexCondition matches condition with regular expression.
type RegexCondition struct {
	PrincipalExclude string `mapstructure:"principal_exclude,omitempty" json:"principal_exclude,omitempty"`
	PrincipalInclude string `mapstructure:"principal_include,omitempty" json:"principal_include,omitempty"`
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
	FromRawJWT *FromRawJWT `mapstructure:"from_raw_jwt,omitempty" json:"from_raw_jwt,omitempty"`
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
	Key    string `mapstructure:"key,omitempty" json:"key,omitempty"`
	Prefix string `mapstructure:"prefix,omitempty" json:"prefix,omitempty"`
	JWKs   *JWKs  `mapstructure:"jwks,omitempty" json:"jwks,omitempty"`
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
	Endpoint string `mapstructure:"endpoint,omitempty" json:"endpoint,omitempty"`
}

// AuditRule is an audit rule to instruct how to audit selected paths/methods.
type AuditRule struct {
	Selector  string `mapstructure:"selector,omitempty" json:"selector,omitempty"`
	Directive string `mapstructure:"directive,omitempty" json:"directive,omitempty"`
	LogType   string `mapstructure:"log_type,omitempty" json:"log_type,omitempty"`
}

// Validate validates the audit rule.
func (r *AuditRule) Validate() error {
	if r.Selector == "" {
		return fmt.Errorf("audit rule selector is empty")
	}
	return nil
}

// SetDefault sets default for the audit rule.
func (r *AuditRule) SetDefault() {
	r.Directive = AuditRuleDirectiveDefault
	r.LogType = AuditLogRequest_DATA_ACCESS.String()
}
