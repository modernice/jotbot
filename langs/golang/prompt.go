package golang

import (
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/modernice/jotbot/generate"
)

// Prompt generates a templated GoDoc comment block based on the provided input,
// which includes the identifier and associated code. It ensures that the
// generated comment adheres to idiomatic GoDoc conventions and instructs users
// on how to write descriptive comments without including technical details such
// as external links or source code examples. The output is designed to guide
// the user in documenting their code effectively while maintaining consistency
// with Go library documentation standards.
func Prompt(input generate.PromptInput) string {
	target := Target(input.Identifier)
	simple := simpleIdentifier(input.Identifier)
	return heredoc.Docf(`
		Write a comment for %s in idiomatic GoDoc format. Do not include any external links, source code, or (code) examples.

		Describe what %s does but not what it _technically_ is. For example, if %s is a function that adds two integers, you must not describe it as a "function that adds two integers." Instead, you must describe it as "adds two integers.".
		
		You must enclose references to other types within brackets ([]). For example, if %q is a function that returns a *Foo, you must describe it as "returns a [*Foo].".

		You must begin the comment exactly with "%s ", and maintain the writing style consistent with Go library documentation.

		Output only the unquoted comment, do not include comment markers (

		Keep the comment as short as possible while still being descriptive.

		Here is the source code for reference:
		---
		# %s
		%s
	`,
		target,
		target,
		simple,
		simple,
		simple,
		input.File,
		input.Code,
	)
}

// Target constructs a string representation of a given identifier within Go
// source code, indicating whether it is a function, type, or variable by
// prefixing the identifier with an appropriate label. If the identifier does
// not match any of the expected formats, it returns the identifier as-is
// enclosed in quotes.
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
