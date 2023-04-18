package generate

func Flatten(files []File) []Generation {
	var all []Generation
	for _, file := range files {
		all = append(all, file.Generations...)
	}
	return all
}
