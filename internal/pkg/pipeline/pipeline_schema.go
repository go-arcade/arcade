package pipeline

// Pipeline represents the top-level pipeline configuration
type Pipeline struct {
	Namespace string            `json:"namespace"`
	Variables map[string]string `json:"variables,omitempty"`
	Jobs      []Job             `json:"jobs"`
}

// Job represents a pipeline job (formerly task)
type Job struct {
	// Job name (required)
	Name string `json:"name"`
	// Job description
	Description string `json:"description,omitempty"`
	// Job environment variables
	Env map[string]string `json:"env,omitempty"`
	// Job concurrency
	Concurrency string `json:"concurrency,omitempty"`
	// Job timeout
	Timeout string `json:"timeout,omitempty"`
	// Job retry
	Retry *Retry `json:"retry,omitempty"`
	// Job when
	When string `json:"when,omitempty"`
	// Job source
	Source *Source `json:"source,omitempty"`
	// Job approval
	Approval *Approval `json:"approval,omitempty"`
	// Job steps
	Steps []Step `json:"steps"`
	// Job target
	Target *Target `json:"target,omitempty"`
	// Job notify
	Notify *Notify `json:"notify,omitempty"`
	// Job triggers
	Triggers []Trigger `json:"triggers,omitempty"`
}

// Retry configuration for job retry logic
type Retry struct {
	// Retry max attempts
	MaxAttempts int `json:"max_attempts"`
	// Retry delay
	Delay string `json:"delay,omitempty"`
}

// Step represents a pipeline step (formerly stage)
type Step struct {
	// Step name (required)
	Name string `json:"name"`
	// Step uses (required)
	Uses string `json:"uses"`
	// Step action
	Action string `json:"action,omitempty"`
	// Step params
	Params map[string]any `json:"params,omitempty"`
	// Step environment variables
	Env map[string]string `json:"env,omitempty"`
	// Step continue on error
	ContinueOnError bool `json:"continue_on_error,omitempty"`
	// Step timeout
	Timeout string `json:"timeout,omitempty"`
	// Step when
	When string `json:"when,omitempty"`
	// Agent selector for step execution
	AgentSelector *AgentSelector `json:"agent_selector,omitempty"`
	// Run on agent (if true, step runs on agent; if false, runs locally)
	RunOnAgent bool `json:"run_on_agent,omitempty"`
}

// Source represents the source configuration (Git / Artifact / S3 / Custom)
type Source struct {
	// Source type (required)
	Type string `json:"type"` // enum: git, artifact, s3, custom
	// Source repository
	Repo string `json:"repo,omitempty"`
	// Source branch
	Branch string `json:"branch,omitempty"`
	// Source authentication
	Auth *SourceAuth `json:"auth,omitempty"`
}

// SourceAuth represents authentication for source
type SourceAuth struct {
	// Source username
	Username string `json:"username,omitempty"`
	// Source password
	Password string `json:"password,omitempty"`
	// Source token
	Token string `json:"token,omitempty"`
}

// Approval represents approval configuration
type Approval struct {
	// Approval required
	Required bool `json:"required"`
	// Approval type
	Type string `json:"type,omitempty"` // enum: manual, auto
	// Approval plugin
	Plugin string `json:"plugin,omitempty"`
	// Approval params
	Params map[string]any `json:"params,omitempty"`
}

// Target represents deployment target configuration
type Target struct {
	// Target type
	Type string `json:"type"` // enum: k8s, vm, docker, s3, custom
	// Target config
	Config map[string]any `json:"config,omitempty"`
}

// Notify represents notification configuration
type Notify struct {
	// Notify on success
	OnSuccess *NotifyItem `json:"on_success,omitempty"`
	// Notify on failure
	OnFailure *NotifyItem `json:"on_failure,omitempty"`
}

// NotifyItem represents a notification item
type NotifyItem struct {
	// Notify plugin
	Plugin string `json:"plugin"`
	// Notify action
	Action string         `json:"action"` // e.g., "Send" or "SendTemplate"
	Params map[string]any `json:"params,omitempty"`
}

// Trigger represents a pipeline trigger
type Trigger struct {
	// Trigger type
	Type string `json:"type"` // enum: manual, cron, event
	// Trigger options
	Options map[string]any `json:"options,omitempty"`
}

// AgentSelector defines criteria for selecting an agent
type AgentSelector struct {
	// MatchLabels requires all specified labels to match (AND logic)
	MatchLabels map[string]string `json:"match_labels,omitempty"`
	// MatchExpressions provides more complex matching rules
	MatchExpressions []LabelExpression `json:"match_expressions,omitempty"`
}

// LabelExpression defines a label matching expression
type LabelExpression struct {
	// Label key
	Key string `json:"key"`
	// Operator: In, NotIn, Exists, NotExists, Gt, Lt
	Operator string `json:"operator"`
	// Values for operator comparison
	Values []string `json:"values,omitempty"`
}
