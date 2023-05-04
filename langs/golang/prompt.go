package golang

import (
	"fmt"

	"github.com/modernice/jotbot/generate"
)

func Prompt(input generate.Input) string {
	return fmt.Sprintf(
		"Write a concise documentation for %s in GoDoc format. Do not output any links. Do not output the source code code or code examples. Use brackets to enclose symbol references. Start the first sentence with %q. Write in the style of Go library documentation. This is the source code of the file:\n\n%s",
		input.Target,
		fmt.Sprintf("%s ", input.Identifier),
		input.Code,
	)
}
