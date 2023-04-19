package generate

import (
	"io/fs"

	"github.com/modernice/opendocs/patch"
)

// Patch generates a patch file for a given repository [fs.FS] and a list of
// [Generation]s. It returns a new [patch.Patch] with the generated patch file.
// Additional options can be passed as variadic arguments.
func Patch(repo fs.FS, gens []Generation, opts ...patch.Option) *patch.Patch {
	p := patch.New(repo, opts...)
	for _, gen := range gens {
		p.Comment(gen.File, gen.Identifier, gen.Doc)
	}
	return p
}
