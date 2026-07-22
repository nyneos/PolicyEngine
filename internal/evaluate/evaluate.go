package evaluate

import (
	"fmt"
	"strconv"
	"strings"

	"PolicyService/internal/model"
)

var actionRank = map[string]int{
	"HardBlock":        4,
	"TriggerApproval":  3,
	"SoftWarning":      2,
	"NotifyOnly":       1,
	"AutoApprove":      0,
}

// Run evaluates each policy independently and aggregates the most restrictive breach action.
func Run(req model.EvaluateRequest) model.EvaluateResponse {
	out := model.EvaluateResponse{Success: true, Results: make([]model.PolicyResult, 0, len(req.Policies))}
	best := ""

	for _, p := range req.Policies {
		res := evalOne(p, req.Variables)
		out.Results = append(out.Results, res)
		if res.Result == "BREACH" && actionRank[res.Action] > actionRank[best] {
			best = res.Action
		}
	}
	out.AggregatedAction = best
	return out
}

func evalOne(p model.PolicySnapshot, vars map[string]string) model.PolicyResult {
	base := model.PolicyResult{PolicyID: p.PolicyID, Code: p.Code}

	var res model.PolicyResult
	switch p.RuleType {
	case "threshold":
		res = evalThreshold(p, vars, base)
	case "list":
		res = evalList(p, vars, base)
	case "formula":
		res = evalFormula(p, vars, base)
	case "slabs", "composition":
		base.Result = "PASS"
		base.Message = fmt.Sprintf("rule_type %s evaluator not fully implemented yet — treated as PASS", p.RuleType)
		res = base
	default:
		base.Result = "ERROR"
		base.Message = "unknown rule_type: " + p.RuleType
		return base
	}

	// Additional Condition (PEL) — AND with main rule when present.
	if res.Result == "PASS" && strings.TrimSpace(p.AddlExpression) != "" {
		ok, err := EvalPELBool(p.AddlExpression, vars)
		if err != nil {
			res.Result = "ERROR"
			res.Message = "addl_expression: " + err.Error()
			res.Action = ""
			return res
		}
		if !ok {
			res.Result = "BREACH"
			res.Action = p.ActionOnBreach
			res.Message = "additional condition failed: " + strings.TrimSpace(p.AddlExpression)
			return res
		}
	}
	return res
}

func evalFormula(p model.PolicySnapshot, vars map[string]string, base model.PolicyResult) model.PolicyResult {
	exprStr := strings.TrimSpace(p.FormulaExpression)
	if exprStr == "" {
		base.Result = "ERROR"
		base.Message = "formula_expression is empty"
		return base
	}
	// Boolean formula: expression must evaluate to bool.
	if strings.EqualFold(p.FormulaReturnType, "boolean") || p.FormulaOperator == "" {
		ok, err := EvalPELBool(exprStr, vars)
		if err != nil {
			base.Result = "ERROR"
			base.Message = "formula: " + err.Error()
			return base
		}
		if ok {
			base.Result = "PASS"
			return base
		}
		base.Result = "BREACH"
		base.Action = p.ActionOnBreach
		base.Message = "formula condition failed"
		return base
	}
	// Numeric formula compared to FormulaValue.
	out, err := EvalPELNumeric(exprStr, vars)
	if err != nil {
		base.Result = "ERROR"
		base.Message = "formula: " + err.Error()
		return base
	}
	pass := compare(out, p.FormulaOperator, p.FormulaValue)
	if pass {
		base.Result = "PASS"
		return base
	}
	base.Result = "BREACH"
	base.Action = p.ActionOnBreach
	base.Message = fmt.Sprintf("formula %s %s %v failed (actual=%v)", exprStr, p.FormulaOperator, p.FormulaValue, out)
	return base
}

func evalThreshold(p model.PolicySnapshot, vars map[string]string, base model.PolicyResult) model.PolicyResult {
	raw, ok := vars[p.ThrVariable]
	if !ok || strings.TrimSpace(raw) == "" {
		return nullResult(p, base, p.ThrVariable)
	}
	val, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		base.Result = "ERROR"
		base.Message = "invalid numeric value for " + p.ThrVariable
		return base
	}
	limit := p.ThrValue
	if p.ThrValueMode == "PercentOf" && p.ThrPercentBase != "" {
		baseRaw, ok := vars[p.ThrPercentBase]
		if !ok {
			return nullResult(p, base, p.ThrPercentBase)
		}
		baseVal, err := strconv.ParseFloat(baseRaw, 64)
		if err != nil || baseVal == 0 {
			base.Result = "ERROR"
			base.Message = "invalid percent base " + p.ThrPercentBase
			return base
		}
		limit = baseVal * (p.ThrValue / 100.0)
	}

	pass := compare(val, p.ThrOperator, limit)
	if pass {
		base.Result = "PASS"
		return base
	}
	base.Result = "BREACH"
	base.Action = p.ActionOnBreach
	base.Message = fmt.Sprintf("%s %s %v failed (actual=%v limit=%v)", p.ThrVariable, p.ThrOperator, limit, val, limit)
	return base
}

func evalList(p model.PolicySnapshot, vars map[string]string, base model.PolicyResult) model.PolicyResult {
	raw, ok := vars[p.ListTargetField]
	if !ok {
		return nullResult(p, base, p.ListTargetField)
	}
	val := raw
	inList := false
	for _, item := range p.ListValues {
		if p.ListCaseSens {
			if item == val {
				inList = true
				break
			}
		} else if strings.EqualFold(item, val) {
			inList = true
			break
		}
	}
	okResult := (p.ListMode == "Include" && inList) || (p.ListMode == "Exclude" && !inList)
	if okResult {
		base.Result = "PASS"
		return base
	}
	base.Result = "BREACH"
	base.Action = p.ActionOnBreach
	base.Message = fmt.Sprintf("%s list %s check failed for value %q", p.ListTargetField, p.ListMode, val)
	return base
}

func nullResult(p model.PolicySnapshot, base model.PolicyResult, missing string) model.PolicyResult {
	switch p.NullHandling {
	case "PassThrough":
		base.Result = "PASS"
		base.Message = "null PassThrough for " + missing
		return base
	case "UseDefault":
		base.Result = "ERROR"
		base.Message = "UseDefault not applied in scaffold for " + missing
		return base
	default: // FailSafe
		base.Result = "BREACH"
		base.Action = "HardBlock"
		base.Message = "FailSafe: missing variable " + missing
		return base
	}
}

func compare(val float64, op string, limit float64) bool {
	switch op {
	case "<":
		return val < limit
	case "<=":
		return val <= limit
	case ">":
		return val > limit
	case ">=":
		return val >= limit
	case "=":
		return val == limit
	case "!=":
		return val != limit
	default:
		return false
	}
}
