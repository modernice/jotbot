package golang_test

import (
	"github.com/modernice/jotbot"
	"github.com/modernice/jotbot/generate"
	"github.com/modernice/jotbot/langs/golang"
	"github.com/modernice/jotbot/patch"
)

var _ interface {
	generate.Language
	patch.Language
	jotbot.Language
} = (*golang.Service)(nil)
