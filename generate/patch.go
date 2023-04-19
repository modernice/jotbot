package generate

import (
	"io/fs"

	"github.com/modernice/opendocs/patch"
)

// Patch returns a new *patch.Patch that generates documentation patches for the
// given repo and Generation slices, with the given patch.Options applied. It
// comments each generation's file, identifier, and doc using p.Comment.
func Patch(repo fs.FS, gens []Generation, opts ...patch.Option) *patch.Patch {
	p := patch.New(repo, opts...)
	for _, gen := range gens {
		p.Comment(gen.File, gen.Identifier, gen.Doc)
	}
	return p
}
