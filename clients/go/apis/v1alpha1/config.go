package v1alpha1

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

// RegexCondition matches condition with regular expression.
type RegexCondition struct {
	PrincipalExclude string `mapstructure:"principal_exclude,omitempty" json:"principal_exclude,omitempty"`
	PrincipalInclude string `mapstructure:"principal_include,omitempty" json:"principal_include,omitempty"`
}

// SecurityContext provides instructive info for where to retrieve
// the security context, e.g. authentication info.
type SecurityContext struct {
	FromRawJWT *FromRawJWT `mapstructure:"from_raw_jwt,omitempty" json:"from_raw_jwt,omitempty"`
}

// FromRawJWT provides info for how to retrieve security context from
// a raw JWT.
type FromRawJWT struct {
	Key    string `mapstructure:"key,omitempty" json:"key,omitempty"`
	Prefix string `mapstructure:"prefix,omitempty" json:"prefix,omitempty"`
	JWKs   *JWKs  `mapstructure:"jwks,omitempty" json:"jwks,omitempty"`
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
