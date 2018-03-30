package main

type stringSorter []string

func (ss stringSorter) Len() int {
	return len(ss)
}

func (ss stringSorter) Less(i, j int) bool {
	return ss[i] < ss[j]
}

func (ss stringSorter) Swap(i, j int) {
	ss[i], ss[j] = ss[j], ss[i]
}
