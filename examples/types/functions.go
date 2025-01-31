package types

type FunctionWithParam func(int)

type FunctionWithNamedParameters func(foo int, bar int)

type FunctionWithSliceParameter func([]int)

type FunctionWithVariadic func(int, ...int)

type FunctionWithReturn func(int) int

type FunctionWithMultipleReturn func(int) (int, int)

type FunctionWithMultipleReturnAndNamedReturn func(int) (a int, b int)

type FunctionAlias = func(int)

type FunctionWithFunctionParameter func(func(int))

type FunctionWithFunctionReturned func() func(int) int

type FuncWithPointerReceiver func(*string)
