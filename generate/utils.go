package generate

func Flatten(files []File) []Documentation {
	var all []Documentation
	for _, file := range files {
		all = append(all, file.Docs...)
	}
	return all
}
