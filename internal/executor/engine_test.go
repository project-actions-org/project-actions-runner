package executor

import (
	"testing"
)

func TestParseOptions(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantOptions map[string]string
		wantArgs    []string
	}{
		{
			name:        "no args",
			args:        []string{},
			wantOptions: map[string]string{},
			wantArgs:    []string{},
		},
		{
			name:        "flags only",
			args:        []string{"--verbose", "--env=production"},
			wantOptions: map[string]string{"verbose": "true", "env": "production"},
			wantArgs:    []string{},
		},
		{
			name:        "flags then positional",
			args:        []string{"--verbose", "foo", "bar", "baz"},
			wantOptions: map[string]string{"verbose": "true"},
			wantArgs:    []string{"foo", "bar", "baz"},
		},
		{
			name:        "positional only",
			args:        []string{"foo", "bar"},
			wantOptions: map[string]string{},
			wantArgs:    []string{"foo", "bar"},
		},
		{
			name:        "explicit double-dash separator",
			args:        []string{"--verbose", "--", "foo", "bar"},
			wantOptions: map[string]string{"verbose": "true"},
			wantArgs:    []string{"foo", "bar"},
		},
		{
			name:        "double-dash separator with no positional",
			args:        []string{"--verbose", "--"},
			wantOptions: map[string]string{"verbose": "true"},
			wantArgs:    []string{},
		},
		{
			name:        "flag with equals value",
			args:        []string{"--env=staging", "script", "arg1"},
			wantOptions: map[string]string{"env": "staging"},
			wantArgs:    []string{"script", "arg1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOpts, gotArgs := parseOptions(tt.args)

			if len(gotOpts) != len(tt.wantOptions) {
				t.Errorf("parseOptions() options len = %d, want %d", len(gotOpts), len(tt.wantOptions))
			}
			for k, v := range tt.wantOptions {
				if gotOpts[k] != v {
					t.Errorf("parseOptions() options[%q] = %q, want %q", k, gotOpts[k], v)
				}
			}

			if len(gotArgs) != len(tt.wantArgs) {
				t.Errorf("parseOptions() args = %v, want %v", gotArgs, tt.wantArgs)
				return
			}
			for i, a := range tt.wantArgs {
				if gotArgs[i] != a {
					t.Errorf("parseOptions() args[%d] = %q, want %q", i, gotArgs[i], a)
				}
			}
		})
	}
}
