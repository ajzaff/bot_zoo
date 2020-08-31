package zoo

func assert(message string, cond bool) {
	if !cond {
		panic(message)
	}
}
