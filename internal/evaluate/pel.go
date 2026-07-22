package evaluate

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

// NestFlatCDM turns dotted CDM keys (investment.fd.principal_amount) into a nested
// map so expr can evaluate investment.fd.principal_amount.
func NestFlatCDM(vars map[string]string) map[string]any {
	root := map[string]any{}
	for k, raw := range vars {
		parts := strings.Split(strings.TrimSpace(k), ".")
		if len(parts) == 0 || parts[0] == "" {
			continue
		}
		cur := root
		for i, p := range parts {
			if i == len(parts)-1 {
				cur[p] = coerceValue(raw)
				continue
			}
			next, ok := cur[p].(map[string]any)
			if !ok {
				next = map[string]any{}
				cur[p] = next
			}
			cur = next
		}
	}
	return root
}

func coerceValue(raw string) any {
	s := strings.TrimSpace(raw)
	if s == "" {
		return nil
	}
	if strings.EqualFold(s, "true") {
		return true
	}
	if strings.EqualFold(s, "false") {
		return false
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return s
}

// BuiltinEnv adds AND/OR/NOT/IF/MIN/MAX helpers used by the Policy Form snippet bar.
func BuiltinEnv() map[string]any {
	return map[string]any{
		"AND": func(args ...bool) bool {
			for _, a := range args {
				if !a {
					return false
				}
			}
			return len(args) > 0
		},
		"OR": func(args ...bool) bool {
			for _, a := range args {
				if a {
					return true
				}
			}
			return false
		},
		"NOT": func(a bool) bool { return !a },
		"IF": func(cond bool, thenVal, elseVal any) any {
			if cond {
				return thenVal
			}
			return elseVal
		},
		"MIN": func(vals ...float64) float64 {
			if len(vals) == 0 {
				return 0
			}
			m := vals[0]
			for _, v := range vals[1:] {
				if v < m {
					m = v
				}
			}
			return m
		},
		"MAX": func(vals ...float64) float64 {
			if len(vals) == 0 {
				return 0
			}
			m := vals[0]
			for _, v := range vals[1:] {
				if v > m {
					m = v
				}
			}
			return m
		},
		"ABS": func(v float64) float64 {
			if v < 0 {
				return -v
			}
			return v
		},
		"TRUE":  true,
		"FALSE": false,
	}
}

func mergeEnv(vars map[string]string) map[string]any {
	env := BuiltinEnv()
	for k, v := range NestFlatCDM(vars) {
		env[k] = v
	}
	return env
}

// CompilePEL compiles a boolean PEL expression with CDM + builtins.
func CompilePEL(expression string, sampleVars map[string]string) (*vm.Program, error) {
	exprStr := strings.TrimSpace(expression)
	if exprStr == "" {
		return nil, fmt.Errorf("expression is empty")
	}
	env := mergeEnv(sampleVars)
	program, err := expr.Compile(exprStr, expr.Env(env), expr.AsBool())
	if err != nil {
		return nil, err
	}
	return program, nil
}

// EvalPELBool evaluates a PEL expression to bool. Empty expression → true (no filter).
func EvalPELBool(expression string, vars map[string]string) (bool, error) {
	exprStr := strings.TrimSpace(expression)
	if exprStr == "" {
		return true, nil
	}
	program, err := CompilePEL(exprStr, vars)
	if err != nil {
		return false, err
	}
	out, err := expr.Run(program, mergeEnv(vars))
	if err != nil {
		return false, err
	}
	b, ok := out.(bool)
	if !ok {
		return false, fmt.Errorf("expression did not return bool (got %T)", out)
	}
	return b, nil
}

func EvalPELNumeric(expression string, vars map[string]string) (float64, error) {
	exprStr := strings.TrimSpace(expression)
	if exprStr == "" {
		return 0, fmt.Errorf("expression is empty")
	}
	env := mergeEnv(vars)
	program, err := expr.Compile(exprStr, expr.Env(env), expr.AsFloat64())
	if err != nil {
		return 0, err
	}
	out, err := expr.Run(program, env)
	if err != nil {
		return 0, err
	}
	switch v := out.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	default:
		return 0, fmt.Errorf("expression did not return number (got %T)", out)
	}
}

// ValidatePEL reports whether the expression compiles as a boolean PEL.
func ValidatePEL(expression string) (bool, string) {
	exprStr := strings.TrimSpace(expression)
	if exprStr == "" {
		return false, "expression is empty"
	}
	sample := map[string]string{
		"investment.fd.principal_amount": "100000",
		"investment.fd.interest_rate":    "7.5",
		"cash.entity_balance":            "1000",
	}
	env := mergeEnv(sample)
	program, err := expr.Compile(exprStr, expr.Env(env), expr.AsBool(), expr.AllowUndefinedVariables())
	if err != nil {
		return false, err.Error()
	}
	_ = program
	return true, "ok (expr-lang)"
}
