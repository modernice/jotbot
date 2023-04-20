package generate

// Flatten returns a flattened slice of [Generation](#Generation) structs 
// generated from the input slice of [File](#File) structs.
func Flatten(files []File) []Generation {
	var all []Generation
	for _, file := range files {
		all = append(all, file.Generations...)
	}
	return all
}
