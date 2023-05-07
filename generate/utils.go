package generate

// Flatten takes a slice of [File] and returns a single slice of [Documentation]
// by concatenating the Documentation slices from each File.
func Flatten(files []File) []Documentation {
	var all []Documentation
	for _, file := range files {
		all = append(all, file.Docs...)
	}
	return all
}
