package ts

import (
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/modernice/jotbot/generate"
)

// Prompt generates a concise TSDoc comment prompt based on the provided Input,
// taking into account the type of identifier (e.g. function, property, method)
// and its context. The generated prompt instructs the user to focus on the
// purpose and usage of the identifier, without including links, source code, or
// code examples.
func Prompt(input generate.PromptInput) string {
	switch extractType(input.Identifier) {
	case "prop", "method":
		return propOrMethodPrompt(input)
	case "func":
		return funcPrompt(input)
	default:
		return defaultPrompt(input)
	}
}

func propOrMethodPrompt(input generate.PromptInput) string {
	target := Target(input.Identifier)
	simple := simpleIdentifier(input.Identifier)
	// owner := extractOwner(input.Identifier)

	return heredoc.Docf(`
		Write a comment for %s in TSDoc format. Do not include any external links, source code, or (code) examples.

		Write the comment in natural language. For example, if %s adds two integers, you must not describe it as a "function that adds two integers." Instead, you must describe it as "%s adds two integers.".
		
		You must enclose references to other types within {@link} references. For example, if %q returns a Foo, you must describe it as "returns a {@link Foo}.".

		You should maintain the writing style consistent with TS library documentation.

		Output only the unquoted comment, do not include comment markers (// or /* */).

		Keep the comment as short as possible while still being descriptive.

		Here is the source code for reference:
		---
		# %s
		%s
	`,
		target,
		simple,
		simple,
		simple,
		input.File,
		input.Code,
	)
}

func funcPrompt(input generate.PromptInput) string {
	target := Target(input.Identifier)
	simple := simpleIdentifier(input.Identifier)

	return heredoc.Docf(`
		Write a comment for %s in TSDoc format. Do not include any external links, source code, or (code) examples.

		Write the comment in natural language. For example, if %s adds two integers, you must not describe it as a "function that adds two integers." Instead, you must describe it as "%s adds two integers.".
		
		You must enclose references to other types within {@link} references. For example, if %q returns a Foo, you must describe it as "returns a {@link Foo}.".

		You should maintain the writing style consistent with TS library documentation.

		Output only the unquoted comment, do not include comment markers (// or /* */).

		Keep the comment as short as possible while still being descriptive.

		Here is the source code for reference:
		---
		# %s
		%s
	`,
		target,
		simple,
		simple,
		simple,
		input.File,
		input.Code,
	)
}

func defaultPrompt(input generate.PromptInput) string {
	target := Target(input.Identifier)
	simple := simpleIdentifier(input.Identifier)

	return heredoc.Docf(`
		Write a comment for %s in TSDoc format. Do not include any external links, source code, or (code) examples.

		Write the comment in natural language. For example, if %s is a function that adds two integers, you must not describe it as a "function that adds two integers." Instead, you must describe it as "%s adds two integers.".
		
		You must enclose references to other types within {@link} references. For example, if %q returns a Foo, you must describe it as "returns a {@link Foo}.".

		You should maintain the writing style consistent with TS library documentation.

		Output only the unquoted comment, do not include comment markers (// or /* */).

		Keep the comment as short as possible while still being descriptive.

		Here is the source code for reference:
		---
		# %s
		%s
	`,
		target,
		simple,
		simple,
		simple,
		input.File,
		input.Code,
	)
}

// Target returns a human-readable representation of the given identifier,
// describing its type and name.
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

// func extractOwner(identifier string) string {
// 	parts := strings.Split(identifier, ":")
// 	if len(parts) != 2 {
// 		return identifier
// 	}
// 	if parts = strings.Split(parts[1], "."); len(parts) == 2 {
// 		return parts[0]
// 	}
// 	return identifier
// }
