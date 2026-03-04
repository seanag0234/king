package king

import (
	"strings"
	"testing"

	"github.com/alecthomas/kong"
)

func TestFishCompletion(t *testing.T) {
	type enumFlag struct {
		Do struct {
			Status *string `enum:"ok,setup,dst" help:"Set status" short:"s"`
		} `cmd:"" help:"do it"`
	}

	type completionFlag struct {
		Do struct {
			Super string `help:"complete this string" completion:"echo bla bloep"`
		} `cmd:"" help:"do it"`
	}

	type fileAction struct {
		Do struct {
			File string `help:"complete this file" completion:"<file>"`
		} `cmd:"" help:"do it"`
	}

	type exportAction struct {
		Do struct {
			Env string `help:"complete env var" completion:"<export>"`
		} `cmd:"" help:"do it"`
	}

	type negatableBool struct {
		Do struct {
			Bool *bool `help:"allow a bool" negatable:""`
		} `cmd:"" help:"do it"`
	}

	type commandAliases struct {
		Do struct{} `cmd:"" aliases:"d" help:"do it"`
	}

	type hiddenCommand struct {
		Visible struct{} `cmd:"" help:"visible cmd"`
		Hidden  struct{} `cmd:"" hidden:"" help:"hidden cmd"`
	}

	type positionalCompletion struct {
		Do struct {
			Volume string `arg:"" help:"Volume" completion:"echo a b c"`
		} `cmd:"" help:"do it"`
	}

	type singleQuoteHelp struct {
		Sq struct{} `cmd:"" help:"it's a test"`
	}

	type dollarHelp struct {
		Ds struct{} `cmd:"" help:"costs $5"`
	}

	type nestedSubcommand struct {
		Outer struct {
			Inner struct {
				Deep struct{} `cmd:"" help:"deep cmd"`
			} `cmd:"" help:"inner cmd"`
		} `cmd:"" help:"outer cmd"`
	}

	type emptyHelp struct {
		Do struct {
			Flag string `help:""`
		} `cmd:"" help:"do it"`
	}

	type completionOverEnum struct {
		Do struct {
			Status string `enum:"ok,setup,dst" default:"ok" help:"Set status" completion:"echo custom"`
		} `cmd:"" help:"do it"`
	}

	type boolNoCompletion struct {
		Do struct {
			Verbose *bool `help:"verbose output"`
		} `cmd:"" help:"do it"`
	}

	type multiplePositionals struct {
		Do struct {
			First  string `arg:"" help:"first arg" completion:"echo 1 2 3"`
			Second string `arg:"" help:"second arg" completion:"echo a b c"`
		} `cmd:"" help:"do it"`
	}

	type leafCommand struct {
		Leaf struct{} `cmd:"" help:"a leaf"`
	}

	type multiAlias struct {
		Do struct{} `cmd:"" aliases:"d,doit" help:"do it"`
	}

	tests := []struct {
		name     string
		cli      any
		contains []string
		excludes []string
	}{
		{
			name:     "enum flag",
			cli:      &enumFlag{},
			contains: []string{"-xa 'ok setup dst'", "-s s", "-l status"},
		},
		{
			name:     "command completion on flag",
			cli:      &completionFlag{},
			contains: []string{"-xa '(echo bla bloep)'"},
		},
		{
			name:     "file action completion",
			cli:      &fileAction{},
			contains: []string{"(__fish_complete_path)"},
		},
		{
			name:     "export action completion",
			cli:      &exportAction{},
			contains: []string{"(set -n)"},
		},
		{
			name:     "negatable bool flag",
			cli:      &negatableBool{},
			contains: []string{"-l bool", "-l no-bool"},
		},
		{
			name:     "command aliases",
			cli:      &commandAliases{},
			contains: []string{"-a do", "-a d"},
		},
		{
			name:     "hidden commands excluded",
			cli:      &hiddenCommand{},
			contains: []string{"-a visible"},
			excludes: []string{"-a hidden"},
		},
		{
			name:     "positional with completion",
			cli:      &positionalCompletion{},
			contains: []string{"-a '(echo a b c)'"},
		},
		{
			name:     "help text with single quote",
			cli:      &singleQuoteHelp{},
			contains: []string{`-d 'it'\''s a test'`},
		},
		{
			name:     "help text with dollar sign",
			cli:      &dollarHelp{},
			contains: []string{`-d 'costs $5'`},
		},
		{
			name: "nested subcommand condition",
			cli:  &nestedSubcommand{},
			contains: []string{
				"__fish_seen_subcommand_from outer; and not __fish_seen_subcommand_from",
				"__fish_seen_subcommand_from inner; and __fish_seen_subcommand_from outer",
			},
		},
		{
			name:     "empty help text no crash",
			cli:      &emptyHelp{},
			excludes: []string{"-d ''"},
		},
		{
			name:     "completion takes precedence over enum",
			cli:      &completionOverEnum{},
			contains: []string{"-xa '(echo custom)'"},
			excludes: []string{"-xa 'ok setup dst'"},
		},
		{
			name:     "bool flag no -x",
			cli:      &boolNoCompletion{},
			excludes: []string{" -x "},
		},
		{
			name: "multiple positional args",
			cli:  &multiplePositionals{},
			contains: []string{
				"-a '(echo 1 2 3)'",
				"-a '(echo a b c)'",
			},
		},
		{
			name:     "leaf command",
			cli:      &leafCommand{},
			contains: []string{"-a leaf"},
		},
		{
			name: "multiple aliases",
			cli:  &multiAlias{},
			contains: []string{
				"-a do",
				"-a d",
				"-a doit",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := kong.Must(tt.cli)
			f := &Fish{}
			f.Completion(parser.Model.Node, "testcli")
			out := string(f.Out())

			for _, want := range tt.contains {
				if !strings.Contains(out, want) {
					t.Errorf("output missing %q\nGot:\n%s", want, out)
				}
			}
			for _, nope := range tt.excludes {
				if strings.Contains(out, nope) {
					t.Errorf("output unexpectedly contains %q\nGot:\n%s", nope, out)
				}
			}
		})
	}
}
