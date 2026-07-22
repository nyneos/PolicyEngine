package model

// EvaluateRequest is sent by CIMPLR Go with Active policies + CDM values.
type EvaluateRequest struct {
	ServiceKey string             `json:"service_key,omitempty"`
	EventCode  string             `json:"event_code"`
	ModuleCode string             `json:"module_code,omitempty"`
	FormID     string             `json:"form_id,omitempty"`
	EntityCode string             `json:"entity_code,omitempty"`
	ActorUser  string             `json:"actor_user_id,omitempty"`
	Variables  map[string]string  `json:"variables"` // CDM name → stringified value
	Policies   []PolicySnapshot   `json:"policies"`
}

// PolicySnapshot is a denormalised policy payload for evaluation (no DB access here).
type PolicySnapshot struct {
	PolicyID       string            `json:"policy_id"`
	Code           string            `json:"code"`
	RuleType       string            `json:"rule_type"`
	ActionOnBreach string            `json:"action_on_breach"`
	NullHandling   string            `json:"null_handling"`
	NullDefault    string            `json:"null_handling_default,omitempty"`
	AddlExpression string            `json:"addl_expression,omitempty"`
	// Threshold
	ThrVariable    string  `json:"thr_variable,omitempty"`
	ThrOperator    string  `json:"thr_operator,omitempty"`
	ThrValue       float64 `json:"thr_value,omitempty"`
	ThrValueMode   string  `json:"thr_value_mode,omitempty"`
	ThrPercentBase string  `json:"thr_percent_base,omitempty"`
	// Slabs
	SlabVariable string     `json:"slab_variable,omitempty"`
	SlabRows     []SlabRow  `json:"slab_rows,omitempty"`
	// Composition
	CompBuckets []CompositionBucket `json:"comp_buckets,omitempty"`
	// List
	ListTargetField string   `json:"list_target_field,omitempty"`
	ListMode        string   `json:"list_mode,omitempty"`
	ListValues      []string `json:"list_values,omitempty"`
	ListCaseSens    bool     `json:"list_case_sensitive,omitempty"`
	// Formula
	FormulaExpression string  `json:"formula_expression,omitempty"`
	FormulaReturnType string  `json:"formula_return_type,omitempty"`
	FormulaOperator   string  `json:"formula_operator,omitempty"`
	FormulaValue      float64 `json:"formula_value,omitempty"`
}

type SlabRow struct {
	From         float64  `json:"from"`
	To           *float64 `json:"to"`
	Mode         string   `json:"mode"`
	Action       string   `json:"action"`
	ApprovalRef  string   `json:"approval_ref,omitempty"`
	Label        string   `json:"label,omitempty"`
}

type CompositionBucket struct {
	Label    string   `json:"label"`
	Variable string   `json:"variable"`
	Min      *float64 `json:"min"`
	Max      *float64 `json:"max"`
}

type PolicyResult struct {
	PolicyID string `json:"policy_id"`
	Code     string `json:"code"`
	Result   string `json:"result"` // PASS | BREACH | ERROR
	Action   string `json:"action,omitempty"`
	Message  string `json:"message,omitempty"`
}

type EvaluateResponse struct {
	Success          bool           `json:"success"`
	AggregatedAction string         `json:"aggregated_action,omitempty"` // HardBlock > TriggerApproval > SoftWarning > NotifyOnly
	Results          []PolicyResult `json:"results"`
	Error            string         `json:"error,omitempty"`
}

type PelValidateRequest struct {
	ServiceKey string `json:"service_key,omitempty"`
	Expression string `json:"expression"`
}

type PelValidateResponse struct {
	Success bool   `json:"success"`
	Valid   bool   `json:"valid"`
	Message string `json:"message,omitempty"`
}
