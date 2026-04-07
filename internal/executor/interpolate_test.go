package executor

import "testing"

func TestInterpolateString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		args    []string
		want    string
		wantErr bool
		errMsg  string
	}{
		{name: "<args> replaced with all args joined by space", input: "echo <args>", args: []string{"hello", "world"}, want: "echo hello world"},
		{name: "<args> with single arg", input: "php <args>", args: []string{"artisan"}, want: "php artisan"},
		{name: "<args> with no args errors", input: "echo <args>", args: []string{}, wantErr: true, errMsg: "argument <args> required but no arguments given"},
		{name: "<args> with nil args errors", input: "<args>", args: nil, wantErr: true, errMsg: "argument <args> required but no arguments given"},
		{name: "repeated <args> token both replaced", input: "echo <args> && echo <args>", args: []string{"hi"}, want: "echo hi && echo hi"},
		{name: "<args.0> replaced with first arg", input: "php <args.0>", args: []string{"artisan", "migrate"}, want: "php artisan"},
		{name: "<args.0> and <args.1> both replaced", input: "cmd <args.0> --flag=<args.1>", args: []string{"script", "value"}, want: "cmd script --flag=value"},
		{name: "<args.2> out of bounds errors", input: "cmd <args.2>", args: []string{"only", "two"}, wantErr: true, errMsg: "argument <args.2> required but only 2 argument(s) given"},
		{name: "<args.0> with no args errors", input: "<args.0>", args: []string{}, wantErr: true, errMsg: "argument <args.0> required but only 0 argument(s) given"},
		{name: "<args.length> returns count as string", input: "echo <args.length> args", args: []string{"a", "b", "c"}, want: "echo 3 args"},
		{name: "<args.length> with zero args returns 0 no error", input: "echo <args.length>", args: []string{}, want: "echo 0"},
		{name: "<args.length> with nil args returns 0 no error", input: "echo <args.length>", args: nil, want: "echo 0"},
		{name: "no tokens passed through unchanged", input: "echo hello world", args: []string{"ignored"}, want: "echo hello world"},
		{name: "non-args angle brackets passed through unchanged", input: "echo <notargs> and <also>", args: []string{"x"}, want: "echo <notargs> and <also>"},
		{name: "mixed tokens in one string", input: "<args.0> <args.1> total=<args.length>", args: []string{"foo", "bar"}, want: "foo bar total=2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := interpolateString(tt.input, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("interpolateString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("interpolateString() error = %q, want %q", err.Error(), tt.errMsg)
				}
				return
			}
			if got != tt.want {
				t.Errorf("interpolateString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInterpolateConfig(t *testing.T) {
	t.Run("interpolates top-level string value", func(t *testing.T) {
		config := map[string]interface{}{"run": "echo <args>"}
		got, err := interpolateConfig(config, []string{"hello"})
		if err != nil { t.Fatalf("unexpected error: %v", err) }
		if got["run"] != "echo hello" { t.Errorf("got %q, want %q", got["run"], "echo hello") }
	})
	t.Run("interpolates nested map values", func(t *testing.T) {
		config := map[string]interface{}{"with": map[string]interface{}{"command": "php <args.0>"}}
		got, err := interpolateConfig(config, []string{"artisan"})
		if err != nil { t.Fatalf("unexpected error: %v", err) }
		nested := got["with"].(map[string]interface{})
		if nested["command"] != "php artisan" { t.Errorf("got %q, want %q", nested["command"], "php artisan") }
	})
	t.Run("interpolates slice of strings", func(t *testing.T) {
		config := map[string]interface{}{"cmds": []interface{}{"echo <args.0>", "echo <args.1>"}}
		got, err := interpolateConfig(config, []string{"foo", "bar"})
		if err != nil { t.Fatalf("unexpected error: %v", err) }
		cmds := got["cmds"].([]interface{})
		if cmds[0] != "echo foo" { t.Errorf("cmds[0] = %q, want %q", cmds[0], "echo foo") }
		if cmds[1] != "echo bar" { t.Errorf("cmds[1] = %q, want %q", cmds[1], "echo bar") }
	})
	t.Run("interpolates map nested inside slice", func(t *testing.T) {
		config := map[string]interface{}{
			"steps": []interface{}{
				map[string]interface{}{"cmd": "echo <args.0>"},
			},
		}
		got, err := interpolateConfig(config, []string{"hello"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		steps := got["steps"].([]interface{})
		step := steps[0].(map[string]interface{})
		if step["cmd"] != "echo hello" {
			t.Errorf("got %q, want %q", step["cmd"], "echo hello")
		}
	})
	t.Run("leaves non-string values unchanged", func(t *testing.T) {
		config := map[string]interface{}{"count": 42, "flag": true}
		got, err := interpolateConfig(config, []string{})
		if err != nil { t.Fatalf("unexpected error: %v", err) }
		if got["count"] != 42 { t.Errorf("count changed: got %v", got["count"]) }
		if got["flag"] != true { t.Errorf("flag changed: got %v", got["flag"]) }
	})
	t.Run("returns error on unresolvable token", func(t *testing.T) {
		config := map[string]interface{}{"run": "<args>"}
		_, err := interpolateConfig(config, []string{})
		if err == nil { t.Error("expected error, got nil") }
	})
	t.Run("empty config returns empty map", func(t *testing.T) {
		got, err := interpolateConfig(map[string]interface{}{}, []string{})
		if err != nil { t.Fatalf("unexpected error: %v", err) }
		if len(got) != 0 { t.Errorf("expected empty map, got %v", got) }
	})
}
