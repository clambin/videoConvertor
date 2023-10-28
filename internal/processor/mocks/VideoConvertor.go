// Code generated by mockery v2.36.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// VideoConvertor is an autogenerated mock type for the VideoConvertor type
type VideoConvertor struct {
	mock.Mock
}

type VideoConvertor_Expecter struct {
	mock *mock.Mock
}

func (_m *VideoConvertor) EXPECT() *VideoConvertor_Expecter {
	return &VideoConvertor_Expecter{mock: &_m.Mock}
}

// Convert provides a mock function with given fields: ctx, input, output, targetCodec
func (_m *VideoConvertor) Convert(ctx context.Context, input string, output string, targetCodec string) error {
	ret := _m.Called(ctx, input, output, targetCodec)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) error); ok {
		r0 = rf(ctx, input, output, targetCodec)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// VideoConvertor_Convert_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Convert'
type VideoConvertor_Convert_Call struct {
	*mock.Call
}

// Convert is a helper method to define mock.On call
//   - ctx context.Context
//   - input string
//   - output string
//   - targetCodec string
func (_e *VideoConvertor_Expecter) Convert(ctx interface{}, input interface{}, output interface{}, targetCodec interface{}) *VideoConvertor_Convert_Call {
	return &VideoConvertor_Convert_Call{Call: _e.mock.On("Convert", ctx, input, output, targetCodec)}
}

func (_c *VideoConvertor_Convert_Call) Run(run func(ctx context.Context, input string, output string, targetCodec string)) *VideoConvertor_Convert_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string), args[3].(string))
	})
	return _c
}

func (_c *VideoConvertor_Convert_Call) Return(_a0 error) *VideoConvertor_Convert_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *VideoConvertor_Convert_Call) RunAndReturn(run func(context.Context, string, string, string) error) *VideoConvertor_Convert_Call {
	_c.Call.Return(run)
	return _c
}

// NewVideoConvertor creates a new instance of VideoConvertor. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewVideoConvertor(t interface {
	mock.TestingT
	Cleanup(func())
}) *VideoConvertor {
	mock := &VideoConvertor{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
