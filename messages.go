package tabcomplete

type Message struct {
	kind interface{}
	id   int
}

type completed struct {
	input      string
	candidates []string
}

type tabErr struct {
	input string
	err   error
}

type clear struct{}

type moveNext struct{}

type movePrev struct{}
