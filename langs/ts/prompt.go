package ts

import (
	"fmt"
	"strings"

	"github.com/modernice/jotbot/generate"
)

func Prompt(input generate.Input) string {
	target := Target(input.Identifier)
	simple := simpleIdentifier(input.Identifier)
	return fmt.Sprintf(
		"Write a concise documentation for %s. Do not include links, source code, or code examples. Write the comment only for %s. You must not describe the type of %s, just describe what it does. Enclose symbol references in {@link} braces. You must adhere to the writing style consistent with TypeScript library documentation. Here is the source code for reference:\n\n%s",
		target,
		target,
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
	case "var":
		return fmt.Sprintf("variable %q", parts[1])
	case "class":
		return fmt.Sprintf("class %q", parts[1])
	case "interface":
		return fmt.Sprintf("interface %q", parts[1])
	case "func":
		return fmt.Sprintf(`function "%s()"`, parts[1])
	case "method":
		return fmt.Sprintf(`method "%s()"`, parts[1])
	case "prop":
		return fmt.Sprintf(`property %q`, parts[1])
	case "type":
		return fmt.Sprintf(`type %q`, parts[1])
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
