package golang

import (
	"fmt"
	"strings"

	"github.com/modernice/jotbot/generate"
)

func Prompt(input generate.Input) string {
	return fmt.Sprintf(
		"Write a concise documentation for %s in GoDoc format. Do not output any links. Do not output the source code code or code examples. Use brackets to enclose symbol references. Start the first sentence with %q. Write in the style of Go library documentation. This is the source code of the file:\n\n%s",
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
		return fmt.Sprintf("function '%s'", parts[1])
	case "type":
		return fmt.Sprintf("type '%s'", parts[1])
	case "var":
		return fmt.Sprintf("variable '%s'", parts[1])
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
