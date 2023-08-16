package interfaceoutput

// Plain interfaces should output the unknown type, which is `any` by default but can also be configured as `unknown`).
type InterfaceOnly interface {
	InterfaceMethod() string
}
