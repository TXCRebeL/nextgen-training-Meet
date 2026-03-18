package stack

type Stack[T any] interface {
	Push(x T) error
	Pop() (T, error)
	Peek() (T, error)
	IsEmpty() bool
	Size() int
}
