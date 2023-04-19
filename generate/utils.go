package generate

// Flatten takes a slice of Files and returns a flattened slice of Generations. 
// It appends all Generations from each File in the input slice to the output 
// slice.
func Flatten(files []File) []Generation {
	var all []Generation
	for _, file := range files {
		all = append(all, file.Generations...)
	}
	return all
}
