package golang

import (
	"fmt"
	"strings"

	"github.com/modernice/jotbot/generate"
)

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
		return fmt.Sprintf("'%s'", identifier)
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
