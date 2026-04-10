package executor

import (
	"fmt"
	"strings"

	"github.com/project-actions/runner/internal/actions"
	"github.com/project-actions/runner/internal/parser"
)

func (e *Engine) executeIfExpr(expr *parser.IfExpr, ctx *actions.ExecutionContext) error {
	result, err := evalIfExpr(expr.Expression, ctx)
	if err != nil {
		return fmt.Errorf("if: expression error: %w", err)
	}

	if result {
		for i := range expr.ThenSteps {
			if err := e.ExecuteStep(&expr.ThenSteps[i], ctx); err != nil {
				return err
			}
		}
	} else {
		for i := range expr.ElseSteps {
			if err := e.ExecuteStep(&expr.ElseSteps[i], ctx); err != nil {
				return err
			}
		}
	}
	return nil
}

// evalIfExpr evaluates a simple boolean DSL expression.
// Supported: option.x, env.X, item.field, ==, !=, &&, ||, !
func evalIfExpr(expr string, ctx *actions.ExecutionContext) (bool, error) {
	expr = strings.TrimSpace(expr)

	// OR (lowest precedence) — split on " || "
	if parts := splitOnOp(expr, "||"); len(parts) > 1 {
		for _, p := range parts {
			v, err := evalIfExpr(p, ctx)
			if err != nil {
				return false, err
			}
			if v {
				return true, nil
			}
		}
		return false, nil
	}

	// AND
	if parts := splitOnOp(expr, "&&"); len(parts) > 1 {
		for _, p := range parts {
			v, err := evalIfExpr(p, ctx)
			if err != nil {
				return false, err
			}
			if !v {
				return false, nil
			}
		}
		return true, nil
	}

	// NOT
	if strings.HasPrefix(expr, "!") {
		v, err := evalIfExpr(strings.TrimSpace(expr[1:]), ctx)
		if err != nil {
			return false, err
		}
		return !v, nil
	}

	// != comparison
	if idx := strings.Index(expr, "!="); idx != -1 {
		lhs := resolveToken(strings.TrimSpace(expr[:idx]), ctx)
		rhsRaw := strings.TrimSpace(expr[idx+2:])
		rhs := resolveRHS(rhsRaw, ctx)
		return lhs != rhs, nil
	}

	// == comparison
	if idx := strings.Index(expr, "=="); idx != -1 {
		lhs := resolveToken(strings.TrimSpace(expr[:idx]), ctx)
		rhsRaw := strings.TrimSpace(expr[idx+2:])
		rhs := resolveRHS(rhsRaw, ctx)
		return lhs == rhs, nil
	}

	// Bare token: truthy if non-empty and not "false"
	val := resolveToken(expr, ctx)
	return val != "" && val != "false", nil
}

// resolveRHS resolves the right-hand side of a comparison.
// If the value is a quoted string literal, the quotes are stripped and the literal is returned.
// Otherwise the value is passed through resolveToken (allowing option.x, env.X, item.field on the RHS).
func resolveRHS(rhs string, ctx *actions.ExecutionContext) string {
	if strings.HasPrefix(rhs, `"`) && strings.HasSuffix(rhs, `"`) {
		return strings.Trim(rhs, `"`)
	}
	return resolveToken(rhs, ctx)
}

// resolveToken looks up a DSL token (option.x, env.X, item.field, varname.field).
func resolveToken(token string, ctx *actions.ExecutionContext) string {
	if strings.HasPrefix(token, "option.") {
		name := token[len("option."):]
		return ctx.Options[name]
	}
	if strings.HasPrefix(token, "env.") {
		name := token[len("env."):]
		return ctx.Environment[name]
	}
	// Loop var: item.field or varname.field
	if dotIdx := strings.Index(token, "."); dotIdx != -1 {
		varName := token[:dotIdx]
		field := token[dotIdx+1:]
		if val, ok := ctx.LoopVars[varName]; ok {
			if m, ok := val.(map[string]interface{}); ok {
				return fmt.Sprint(m[field])
			}
			// var found but is a scalar — field access on scalar returns empty (falsy)
			ctx.Logger.Debug("if: field access on scalar loop variable %q ignored", field)
			return ""
		}
	} else if val, ok := ctx.LoopVars[token]; ok {
		return fmt.Sprint(val)
	}
	return token // treat as string literal for comparison RHS
}

// splitOnOp splits expr on op (surrounded by spaces) without splitting inside sub-expressions.
// This simple implementation handles the flat cases in our DSL.
func splitOnOp(expr, op string) []string {
	padded := " " + op + " "
	if !strings.Contains(expr, padded) {
		return []string{expr}
	}
	parts := strings.Split(expr, padded)
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	return parts
}
