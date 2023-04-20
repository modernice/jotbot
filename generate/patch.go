package generate

import (
	"io/fs"

	"github.com/modernice/jotbot/patch"
)

// Patch creates a new [Patch] that applies [Generation]s to the specified file
// system. It takes a file system object, a slice of [Generation]s, and optional
// [Option]s as input. It returns a pointer to the created [Patch].
func Patch(repo fs.FS, gens []Generation, opts ...patch.Option) *patch.Patch {
	p := patch.New(repo, opts...)
	for _, gen := range gens {
		p.Comment(gen.File, gen.Identifier, gen.Doc)
	}
	return p
}
