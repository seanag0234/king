package king

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/alecthomas/kong"
)

//go:embed fish.tmpl
var fishTemplate string

// Fish is a fish shell completion generator.
type Fish struct {
	name       string
	completion []byte
	Flags      []*kong.Flag // Any global flags that the should Application Node have.
}

func (f *Fish) Out() []byte { return f.completion }

func (f *Fish) Write(w ...io.Writer) error {
	if f.completion == nil {
		return fmt.Errorf("no completion")
	}
	if len(w) > 0 {
		w[0].Write(f.completion)
	}
	return os.WriteFile(f.name+".fish", f.completion, 0644) // no idea what fish needs
}

func (f *Fish) Completion(k *kong.Node, altname string) {
	k.Flags = append(k.Flags, f.Flags...)

	if altname == "" {
		f.name = k.Name
	} else {
		f.name = altname
		k.Name = altname
	}

	funcMap := template.FuncMap{
		"rootName":            func() string { return f.name },
		"path":                func(cmd *kong.Node) string { return cmd.Path() },
		"subcommandCondition": fishSubcommandCondition,
		"flagLines":           func(cmd *kong.Node) []string { return fishFlagLines(f.name, cmd) },
		"visibleChildren":     fishVisibleChildren,
	}

	tmpl := template.Must(template.New("fish").Funcs(funcMap).Parse(fishTemplate))
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, k); err != nil {
		panic(fmt.Sprintf("fish template execution failed: %s", err))
	}
	f.completion = buf.Bytes()
}

// fishFlagLines returns complete fish completion lines for all visible flags of cmd.
func fishFlagLines(rootName string, cmd *kong.Node) []string {
	var lines []string
	for _, f := range cmd.Flags {
		if f.Hidden {
			continue
		}
		var b strings.Builder
		if cmd.Parent == nil {
			fmt.Fprintf(&b, "complete -c %s -f", rootName)
		} else {
			fmt.Fprintf(&b, "complete -c %s -f -n '__fish_seen_subcommand_from %s'", rootName, cmd.Name)
		}
		if !f.IsBool() {
			enums := flagEnums(f)
			if len(enums) > 0 {
				fmt.Fprintf(&b, " -xa '%s'", strings.Join(enums, " "))
			} else {
				b.WriteString(" -x")
			}
		}
		if f.Short != 0 {
			fmt.Fprintf(&b, " -s %c", f.Short)
		}
		fmt.Fprintf(&b, " -l %s", f.Name)
		fmt.Fprintf(&b, " -d \"%s\"", f.Help)
		lines = append(lines, b.String())
	}
	return lines
}

// fishVisibleChildren returns non-nil, non-hidden children of cmd.
func fishVisibleChildren(cmd *kong.Node) []*kong.Node {
	var children []*kong.Node
	for _, c := range cmd.Children {
		if c != nil && !c.Hidden {
			children = append(children, c)
		}
	}
	return children
}

// fishSubcommandCondition returns the fish condition (-n) for when a
// subcommand's completions should appear. Top-level subcommands use
// __fish_use_subcommand. Nested subcommands require all ancestors to be
// present on the command line and no sibling to be chosen yet.
func fishSubcommandCondition(cmd *kong.Node) string {
	if cmd.Parent.Parent == nil {
		return "__fish_use_subcommand"
	}

	ancestors := fishAncestorNames(cmd)
	siblings := fishSiblingNames(cmd)

	// "seen parent; and seen grandparent; ... ; and not seen any-sibling"
	var parts []string
	for _, name := range ancestors {
		parts = append(parts, "__fish_seen_subcommand_from "+name)
	}
	parts = append(parts, "not __fish_seen_subcommand_from "+strings.Join(siblings, " "))

	return strings.Join(parts, "; and ")
}

// fishAncestorNames collects the names of all ancestor subcommands, from
// parent up to (but not including) root.
func fishAncestorNames(cmd *kong.Node) []string {
	var names []string
	for node := cmd.Parent; node.Parent != nil; node = node.Parent {
		names = append(names, node.Name)
	}
	return names
}

// fishSiblingNames collects the names of all visible sibling subcommands
// (children of cmd's parent, including cmd itself).
func fishSiblingNames(cmd *kong.Node) []string {
	var names []string
	for _, c := range cmd.Parent.Children {
		if c != nil && !c.Hidden {
			names = append(names, c.Name)
		}
	}
	return names
}
