package ts

import (
	"fmt"
	"strings"

	"github.com/modernice/jotbot/generate"
)

func Prompt(input generate.Input) string {
	switch extractType(input.Identifier) {
	case "prop", "method":
		return propOrMethodPrompt(input)
	case "func":
		return funcPrompt(input)
	default:
		return defaultPrompt(input)
	}
}

func propOrMethodPrompt(input generate.Input) string {
	target := Target(input.Identifier)
	simple := simpleIdentifier(input.Identifier)
	owner := extractOwner(input.Identifier)

	return fmt.Sprintf(
		"Write a concise TSDoc comment for %s, focusing on its purpose and how it is used within the context of %q. You must not output links, source code, or code examples. Write the comment only for %s. Don't describe what %s is, but what it does. Enclose symbol references in {@link} braces. Here is the source code for reference:\n\n%s",
		target,
		owner,
		target,
		simple,
		input.Code,
	)
}

func funcPrompt(input generate.Input) string {
	target := Target(input.Identifier)
	simple := simpleIdentifier(input.Identifier)

	return fmt.Sprintf(
		"Write a concise TSDoc comment for %s. You must not output links, source code, or code examples. Write the comment only for %s. Don't describe what %s is, but what it does. Enclose symbol references in {@link} braces. Here is the source code for reference:\n\n%s",
		target,
		target,
		simple,
		input.Code,
	)
}

func defaultPrompt(input generate.Input) string {
	target := Target(input.Identifier)
	simple := simpleIdentifier(input.Identifier)

	var beginWith string
	switch extractType(input.Identifier) {
	case "type", "iface", "class":
		beginWith = fmt.Sprintf(`You must begin the comment with "%s ". `, simple)
	}

	return fmt.Sprintf(
		"Write a concise TSDoc comment for %s. You must not output links, source code, or code examples. Write the comment only for %s. Don't describe what %s is, but what it does. Enclose symbol references in {@link} braces. %sHere is the source code for reference:\n\n%s",
		target,
		target,
		simple,
		beginWith,
		input.Code,
	)
}

func Target(identifier string) string {
	parts := strings.Split(identifier, ":")
	if len(parts) != 2 {
		return identifier
	}

	typ := parts[0]
	path := parts[1]
	name := path

	var owner string
	if parts = strings.Split(path, "."); len(parts) == 2 {
		owner = parts[0]
		name = parts[1]
	}

	switch typ {
	case "var":
		return fmt.Sprintf("variable %q", name)
	case "class":
		return fmt.Sprintf("class %q", name)
	case "interface":
		return fmt.Sprintf("interface %q", name)
	case "func":
		return fmt.Sprintf(`function "%s()"`, name)
	case "method":
		return fmt.Sprintf(`method %q of %q"`, name, owner)
	case "prop":
		return fmt.Sprintf(`property %q of %q`, name, owner)
	case "type":
		return fmt.Sprintf(`type %q`, name)
	default:
		return identifier
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

func extractType(identifier string) string {
	parts := strings.Split(identifier, ":")
	if len(parts) == 2 {
		return parts[0]
	}
	return ""
}

func extractOwner(identifier string) string {
	parts := strings.Split(identifier, ":")
	if len(parts) != 2 {
		return identifier
	}
	if parts = strings.Split(parts[1], "."); len(parts) == 2 {
		return parts[0]
	}
	return identifier
}
