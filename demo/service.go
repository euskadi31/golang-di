package demo

// A struct
type A struct {
}

// NewA constructor
// @Service
func NewA() *A {
	return &A{}
}

// B struct
type B struct {
	a *A
}

// NewB constructor
// @Service
func NewB(a *A) *B {
	return &B{
		a: a,
	}
}
