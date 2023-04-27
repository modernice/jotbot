package generate

// // Patch creates a new [Patch] that applies [Generation]s to the specified file
// // system. It takes a file system object, a slice of [Generation]s, and optional
// // [Option]s as input. It returns a pointer to the created [Patch].
// func Patch(repo fs.FS, gens []Generation, opts ...golang.PatchOption) *golang.Patch {
// 	p := golang.NewPatch(repo, opts...)
// 	for _, gen := range gens {
// 		p.Comment(gen.File, gen.Identifier, gen.Doc)
// 	}
// 	return p
// }
