package executor

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var argsPattern = regexp.MustCompile(`<args(?:\.\d+|\.length)?>`)

func interpolateString(s string, args []string) (string, error) {
	var firstErr error

	result := argsPattern.ReplaceAllStringFunc(s, func(token string) string {
		if firstErr != nil {
			return token
		}

		// token is one of: <args>, <args.0>, <args.1>, ..., <args.length>
		inner := token[1 : len(token)-1] // strip < and >

		if inner == "args" {
			if len(args) == 0 {
				firstErr = fmt.Errorf("argument <args> required but no arguments given")
				return token
			}
			return strings.Join(args, " ")
		}

		suffix := inner[len("args"):] // e.g. ".0", ".length"

		if suffix == ".length" {
			return strconv.Itoa(len(args))
		}

		// suffix is ".N"
		indexStr := suffix[1:] // strip leading "."
		// indexStr is guaranteed to be digits-only by the regex, so Atoi cannot fail.
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
	return result, nil
}

func interpolateValue(v interface{}, args []string) (interface{}, error) {
	switch val := v.(type) {
	case string:
		return interpolateString(val, args)
	case map[string]interface{}:
		return interpolateConfig(val, args)
	case []interface{}:
		result := make([]interface{}, len(val))
		for i, elem := range val {
			interpolated, err := interpolateValue(elem, args)
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

func interpolateConfig(config map[string]interface{}, args []string) (map[string]interface{}, error) {
	result := make(map[string]interface{}, len(config))
	for k, v := range config {
		interpolated, err := interpolateValue(v, args)
		if err != nil {
			return nil, err
		}
		result[k] = interpolated
	}
	return result, nil
}
