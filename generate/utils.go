package generate

// Flatten takes a slice of [File] and consolidates all the contained
// [Documentation] into a single slice. It iterates over the files, appending
// each file's documentation to a cumulative slice, which is then returned.
func Flatten(files []File) []Documentation {
	var all []Documentation
	for _, file := range files {
		all = append(all, file.Docs...)
	}
	return all
}
