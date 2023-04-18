package generate

import (
	"io/fs"

	"github.com/modernice/opendocs/patch"
)

func Patch(repo fs.FS, gens []Generation, opts ...patch.Option) *patch.Patch {
	p := patch.New(repo, opts...)
	for _, gen := range gens {
		p.Comment(gen.File, gen.Identifier, gen.Doc)
	}
	return p
}
