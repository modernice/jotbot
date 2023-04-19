package generate

// Flatten takes in a slice of [File] and returns a slice of [Generation]. It 
// flattens the slice of files by appending all the generations from each file 
// into a single slice.
func Flatten(files []File) []Generation {
	var all []Generation
	for _, file := range files {
		all = append(all, file.Generations...)
	}
	return all
}
