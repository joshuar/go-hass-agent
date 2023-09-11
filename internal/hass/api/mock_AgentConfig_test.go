// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package api

import (
	"sync"
)

// Ensure, that AgentConfigMock does implement AgentConfig.
// If this is not the case, regenerate this file with moq.
var _ AgentConfig = &AgentConfigMock{}

// AgentConfigMock is a mock implementation of AgentConfig.
//
//	func TestSomethingThatUsesAgentConfig(t *testing.T) {
//
//		// make and configure a mocked AgentConfig
//		mockedAgentConfig := &AgentConfigMock{
//			GetFunc: func(s string, ifaceVal interface{}) error {
//				panic("mock out the Get method")
//			},
//		}
//
//		// use mockedAgentConfig in code that requires AgentConfig
//		// and then make assertions.
//
//	}
type AgentConfigMock struct {
	// GetFunc mocks the Get method.
	GetFunc func(s string, ifaceVal interface{}) error

	// calls tracks calls to the methods.
	calls struct {
		// Get holds details about calls to the Get method.
		Get []struct {
			// S is the s argument value.
			S string
			// IfaceVal is the ifaceVal argument value.
			IfaceVal interface{}
		}
	}
	lockGet sync.RWMutex
}

// Get calls GetFunc.
func (mock *AgentConfigMock) Get(s string, ifaceVal interface{}) error {
	if mock.GetFunc == nil {
		panic("AgentConfigMock.GetFunc: method is nil but AgentConfig.Get was just called")
	}
	callInfo := struct {
		S        string
		IfaceVal interface{}
	}{
		S:        s,
		IfaceVal: ifaceVal,
	}
	mock.lockGet.Lock()
	mock.calls.Get = append(mock.calls.Get, callInfo)
	mock.lockGet.Unlock()
	return mock.GetFunc(s, ifaceVal)
}

// GetCalls gets all the calls that were made to Get.
// Check the length with:
//
//	len(mockedAgentConfig.GetCalls())
func (mock *AgentConfigMock) GetCalls() []struct {
	S        string
	IfaceVal interface{}
} {
	var calls []struct {
		S        string
		IfaceVal interface{}
	}
	mock.lockGet.RLock()
	calls = mock.calls.Get
	mock.lockGet.RUnlock()
	return calls
}