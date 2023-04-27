package golang

// var _ jotbot.Language[*Finder, *Patch] = (*Go)(nil)

type Go struct{}

func (lang *Go) Finder() *Finder {
	return nil
}

func (lang *Go) Patcher() *Patch {
	return nil
}
