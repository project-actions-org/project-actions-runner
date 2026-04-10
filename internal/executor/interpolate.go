package executor

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var argsPattern = regexp.MustCompile(`<args(?:\.\d+|\.length)?>`)
var loopVarPattern = regexp.MustCompile(`<([a-zA-Z_][a-zA-Z0-9_]*)(?:\.([a-zA-Z0-9_]+))?>`)

func interpolateString(s string, args []string, loopVars map[string]interface{}) (string, error) {
	// Step 1: replace <args>, <args.N>, <args.length> tokens
	var firstErr error
	result := argsPattern.ReplaceAllStringFunc(s, func(token string) string {
		if firstErr != nil {
			return token
		}
		inner := token[1 : len(token)-1]
		if inner == "args" {
			if len(args) == 0 {
				firstErr = fmt.Errorf("argument <args> required but no arguments given")
				return token
			}
			return strings.Join(args, " ")
		}
		suffix := inner[len("args"):]
		if suffix == ".length" {
			return strconv.Itoa(len(args))
		}
		indexStr := suffix[1:]
		index, _ := strconv.Atoi(indexStr)
		if index >= len(args) {
			firstErr = fmt.Errorf("argument %s required but only %d argument(s) given", token, len(args))
			return token
		}
		return args[index]
	})
	if firstErr != nil {
		return "", firstErr
	}

	// Step 2: replace <item>, <item.field>, <varname>, <varname.field> tokens
	if len(loopVars) > 0 {
		result = loopVarPattern.ReplaceAllStringFunc(result, func(token string) string {
			inner := token[1 : len(token)-1]
			dotIdx := strings.Index(inner, ".")
			var varName, field string
			if dotIdx == -1 {
				varName = inner
			} else {
				varName = inner[:dotIdx]
				field = inner[dotIdx+1:]
			}
			val, ok := loopVars[varName]
			if !ok {
				return token // not a loop var — leave unchanged
			}
			if field == "" {
				return fmt.Sprint(val)
			}
			if m, ok := val.(map[string]interface{}); ok {
				if fv, ok := m[field]; ok {
					return fmt.Sprint(fv)
				}
			}
			return token
		})
	}

	return result, nil
}

func interpolateValue(v interface{}, args []string, loopVars map[string]interface{}) (interface{}, error) {
	switch val := v.(type) {
	case string:
		return interpolateString(val, args, loopVars)
	case map[string]interface{}:
		return interpolateConfig(val, args, loopVars)
	case []interface{}:
		result := make([]interface{}, len(val))
		for i, elem := range val {
			interpolated, err := interpolateValue(elem, args, loopVars)
			if err != nil {
				return nil, err
			}
			result[i] = interpolated
		}
		return result, nil
	default:
		return v, nil
	}
}

func interpolateConfig(config map[string]interface{}, args []string, loopVars map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{}, len(config))
	for k, v := range config {
		interpolated, err := interpolateValue(v, args, loopVars)
		if err != nil {
			return nil, err
		}
		result[k] = interpolated
	}
	return result, nil
}
