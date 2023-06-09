// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package hass

import (
	"sync"
)

// Ensure, that RegistrationInfoMock does implement RegistrationInfo.
// If this is not the case, regenerate this file with moq.
var _ RegistrationInfo = &RegistrationInfoMock{}

// RegistrationInfoMock is a mock implementation of RegistrationInfo.
//
//	func TestSomethingThatUsesRegistrationInfo(t *testing.T) {
//
//		// make and configure a mocked RegistrationInfo
//		mockedRegistrationInfo := &RegistrationInfoMock{
//			ServerFunc: func() string {
//				panic("mock out the Server method")
//			},
//			TokenFunc: func() string {
//				panic("mock out the Token method")
//			},
//		}
//
//		// use mockedRegistrationInfo in code that requires RegistrationInfo
//		// and then make assertions.
//
//	}
type RegistrationInfoMock struct {
	// ServerFunc mocks the Server method.
	ServerFunc func() string

	// TokenFunc mocks the Token method.
	TokenFunc func() string

	// calls tracks calls to the methods.
	calls struct {
		// Server holds details about calls to the Server method.
		Server []struct {
		}
		// Token holds details about calls to the Token method.
		Token []struct {
		}
	}
	lockServer sync.RWMutex
	lockToken  sync.RWMutex
}

// Server calls ServerFunc.
func (mock *RegistrationInfoMock) Server() string {
	if mock.ServerFunc == nil {
		panic("RegistrationInfoMock.ServerFunc: method is nil but RegistrationInfo.Server was just called")
	}
	callInfo := struct {
	}{}
	mock.lockServer.Lock()
	mock.calls.Server = append(mock.calls.Server, callInfo)
	mock.lockServer.Unlock()
	return mock.ServerFunc()
}

// ServerCalls gets all the calls that were made to Server.
// Check the length with:
//
//	len(mockedRegistrationInfo.ServerCalls())
func (mock *RegistrationInfoMock) ServerCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockServer.RLock()
	calls = mock.calls.Server
	mock.lockServer.RUnlock()
	return calls
}

// Token calls TokenFunc.
func (mock *RegistrationInfoMock) Token() string {
	if mock.TokenFunc == nil {
		panic("RegistrationInfoMock.TokenFunc: method is nil but RegistrationInfo.Token was just called")
	}
	callInfo := struct {
	}{}
	mock.lockToken.Lock()
	mock.calls.Token = append(mock.calls.Token, callInfo)
	mock.lockToken.Unlock()
	return mock.TokenFunc()
}

// TokenCalls gets all the calls that were made to Token.
// Check the length with:
//
//	len(mockedRegistrationInfo.TokenCalls())
func (mock *RegistrationInfoMock) TokenCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockToken.RLock()
	calls = mock.calls.Token
	mock.lockToken.RUnlock()
	return calls
}
