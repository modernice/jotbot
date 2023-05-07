package golang

import (
	"fmt"
	"strings"

	"github.com/modernice/jotbot/generate"
)

// Prompt generates a formatted string instructing the user to write a concise
// GoDoc comment for a given identifier, adhering to Go library documentation
// style and conventions, without including links, source code, or code
// examples. It takes an input of type [generate.Input] and returns the
// generated prompt string.
func Prompt(input generate.Input) string {
	target := Target(input.Identifier)
	simple := simpleIdentifier(input.Identifier)
	return fmt.Sprintf(
		"Write a concise comment for %s in GoDoc format. Do not including links, source code, or code examples. Write the comment only for %s. You must not describe the type of %s, just describe what it does. You must enclose references to other types within brackets ([]). You must begin the comment with %q, and maintain the writing style consistent with Go library documentation. Here is the source code for reference:\n\n%s",
		target,
		target,
		simple,
		simple,
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
