package golang

import (
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/modernice/jotbot/generate"
)

// Prompt takes an input of type [generate.Input] and generates a string that
// forms a GoDoc-style comment for the input. It first creates a target
// description of the identifier contained in the input using the Target
// function, and obtains a simplified identifier with the simpleIdentifier
// function. The resulting string is formatted to provide instructions on how to
// write a GoDoc comment for the targeted identifier.
func Prompt(input generate.PromptInput) string {
	target := Target(input.Identifier)
	simple := simpleIdentifier(input.Identifier)
	return heredoc.Docf(`
		Write a comment for %s in idiomatic GoDoc format. Do not include any external links or source code.
		If you provide a code example, it must be short, concise, and only for %s.

		Describe what %s does but not what it _technically_ is. For example, if %s is a function that adds two integers, you must not describe it as a "function that adds two integers." Instead, you must describe it as "adds two integers.".
		
		You must enclose references to other types within brackets ([]). For example, if %q is a function that returns a *Foo, you must describe it as "returns a [*Foo].".

		You must begin the comment exactly with "%s ", and maintain the writing style consistent with Go library documentation.

		Output only the unquoted comment, do not include comment markers (// or /* */).

		Here is the source code for reference:
		---
		# %s
		%s
	`,
		target,
		target,
		target,
		simple,
		simple,
		simple,
		input.File,
		input.Code,
	)
}

// Target returns a human-readable description of the given identifier, which
// can be a function, type, or variable. It formats the identifier based on its
// kind (func, type, or var) and name.
func Target(identifier string) string {
	parts := strings.Split(identifier, ":")
	if len(parts) != 2 {
		return identifier
	}

	switch parts[0] {
	case "func":
		return fmt.Sprintf(`function "%s()"`, parts[1])
	case "type":
		return fmt.Sprintf("type %q", parts[1])
	case "var":
		return fmt.Sprintf("variable %q", parts[1])
	default:
		return fmt.Sprintf("%q", identifier)
	}
}

func simpleIdentifier(identifier string) string {
	parts := strings.Split(identifier, ":")
	if len(parts) == 2 {
		return removeOwner(parts[1])
	}
	return removeOwner(identifier)
}

func removeOwner(identifier string) string {
	parts := strings.Split(identifier, ".")
	if len(parts) == 2 {
		return parts[1]
	}
	return identifier
}
