package golang

import (
	"fmt"
	"strings"

	"github.com/modernice/jotbot/generate"
)

func Prompt(input generate.Input) string {
	return fmt.Sprintf(
		"Write comprehensive and concise documentation for %s in idiomatic GoDoc format. Avoid including links, source code, or code examples. Enclose symbol references within brackets ([]). Begin the first sentence with %q, and maintain the writing style consistent with Go library documentation. Here is the source code of the file for reference:\n\n%s",
		Target(input.Identifier),
		fmt.Sprintf("%s ", simpleIdentifier(input.Identifier)),
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
