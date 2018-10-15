package demo

// ServiceA struct
type ServiceA struct {
}

// NewServiceA constructor
// @Service
func NewServiceA() *ServiceA {
	return &ServiceA{}
}

// ServiceB struct
type ServiceB struct {
	a *ServiceA
}

// NewServiceB constructor
// @Service
func NewServiceB(a *ServiceA) *ServiceB {
	return &ServiceB{
		a: a,
	}
}

// ServiceC struct
type ServiceC struct {
	b ServiceB
}

// NewServiceC constructor
// @Service
func NewServiceC(b ServiceB) *ServiceC {
	return &ServiceC{
		b: b,
	}
}
