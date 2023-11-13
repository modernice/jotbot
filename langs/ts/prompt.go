package ts

import (
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/modernice/jotbot/generate"
)

// Prompt constructs a TSDoc comment template based on the provided input which
// includes the identifier and source code context. It determines the type of
// the identifier, such as a property, method, or function, and formats a prompt
// accordingly. The generated prompt instructs the user to write a natural
// language description of the identified code element without using technical
// jargon or including extraneous information such as external links or code
// examples. References to other types within the comment should be enclosed
// using {@link} syntax, and the style should align with typical TypeScript
// library documentation conventions.
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

	return heredoc.Docf(`
		Write a comment for %s in TSDoc format. Do not include any external links, source code, or (code) examples.

		Write the comment in natural language. For example, if %s adds two integers, you must not describe it as a "function that adds two integers." Instead, you must describe it as "%s adds two integers.".
		
		You must enclose references to other types within {@link} references. For example, if %q returns a Foo, you must describe it as "returns a {@link Foo}.".

		You should maintain the writing style consistent with TS library documentation.

		Output only the unquoted comment, do not include comment markers (

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

		Output only the unquoted comment, do not include comment markers (

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

		Output only the unquoted comment, do not include comment markers (

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

// Target constructs a descriptive string for an identifier by categorizing it
// and appending relevant information based on its type, such as the name of a
// class, the signature of a function, or the association of a method or
// property with its owner. It handles various identifier types including
// variables, classes, interfaces, functions, methods, properties, and custom
// types. If the identifier does not conform to expected patterns or types, it
// is returned as-is.
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
