// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package sensor

import (
	"sync"
)

// Ensure, that RegistryMock does implement Registry.
// If this is not the case, regenerate this file with moq.
var _ Registry = &RegistryMock{}

// RegistryMock is a mock implementation of Registry.
//
//	func TestSomethingThatUsesRegistry(t *testing.T) {
//
//		// make and configure a mocked Registry
//		mockedRegistry := &RegistryMock{
//			IsDisabledFunc: func(sensor string) bool {
//				panic("mock out the IsDisabled method")
//			},
//			IsRegisteredFunc: func(sensor string) bool {
//				panic("mock out the IsRegistered method")
//			},
//			SetDisabledFunc: func(sensor string, state bool) error {
//				panic("mock out the SetDisabled method")
//			},
//			SetRegisteredFunc: func(sensor string, state bool) error {
//				panic("mock out the SetRegistered method")
//			},
//		}
//
//		// use mockedRegistry in code that requires Registry
//		// and then make assertions.
//
//	}
type RegistryMock struct {
	// IsDisabledFunc mocks the IsDisabled method.
	IsDisabledFunc func(sensor string) bool

	// IsRegisteredFunc mocks the IsRegistered method.
	IsRegisteredFunc func(sensor string) bool

	// SetDisabledFunc mocks the SetDisabled method.
	SetDisabledFunc func(sensor string, state bool) error

	// SetRegisteredFunc mocks the SetRegistered method.
	SetRegisteredFunc func(sensor string, state bool) error

	// calls tracks calls to the methods.
	calls struct {
		// IsDisabled holds details about calls to the IsDisabled method.
		IsDisabled []struct {
			// Sensor is the sensor argument value.
			Sensor string
		}
		// IsRegistered holds details about calls to the IsRegistered method.
		IsRegistered []struct {
			// Sensor is the sensor argument value.
			Sensor string
		}
		// SetDisabled holds details about calls to the SetDisabled method.
		SetDisabled []struct {
			// Sensor is the sensor argument value.
			Sensor string
			// State is the state argument value.
			State bool
		}
		// SetRegistered holds details about calls to the SetRegistered method.
		SetRegistered []struct {
			// Sensor is the sensor argument value.
			Sensor string
			// State is the state argument value.
			State bool
		}
	}
	lockIsDisabled    sync.RWMutex
	lockIsRegistered  sync.RWMutex
	lockSetDisabled   sync.RWMutex
	lockSetRegistered sync.RWMutex
}

// IsDisabled calls IsDisabledFunc.
func (mock *RegistryMock) IsDisabled(sensor string) bool {
	if mock.IsDisabledFunc == nil {
		panic("RegistryMock.IsDisabledFunc: method is nil but Registry.IsDisabled was just called")
	}
	callInfo := struct {
		Sensor string
	}{
		Sensor: sensor,
	}
	mock.lockIsDisabled.Lock()
	mock.calls.IsDisabled = append(mock.calls.IsDisabled, callInfo)
	mock.lockIsDisabled.Unlock()
	return mock.IsDisabledFunc(sensor)
}

// IsDisabledCalls gets all the calls that were made to IsDisabled.
// Check the length with:
//
//	len(mockedRegistry.IsDisabledCalls())
func (mock *RegistryMock) IsDisabledCalls() []struct {
	Sensor string
} {
	var calls []struct {
		Sensor string
	}
	mock.lockIsDisabled.RLock()
	calls = mock.calls.IsDisabled
	mock.lockIsDisabled.RUnlock()
	return calls
}

// IsRegistered calls IsRegisteredFunc.
func (mock *RegistryMock) IsRegistered(sensor string) bool {
	if mock.IsRegisteredFunc == nil {
		panic("RegistryMock.IsRegisteredFunc: method is nil but Registry.IsRegistered was just called")
	}
	callInfo := struct {
		Sensor string
	}{
		Sensor: sensor,
	}
	mock.lockIsRegistered.Lock()
	mock.calls.IsRegistered = append(mock.calls.IsRegistered, callInfo)
	mock.lockIsRegistered.Unlock()
	return mock.IsRegisteredFunc(sensor)
}

// IsRegisteredCalls gets all the calls that were made to IsRegistered.
// Check the length with:
//
//	len(mockedRegistry.IsRegisteredCalls())
func (mock *RegistryMock) IsRegisteredCalls() []struct {
	Sensor string
} {
	var calls []struct {
		Sensor string
	}
	mock.lockIsRegistered.RLock()
	calls = mock.calls.IsRegistered
	mock.lockIsRegistered.RUnlock()
	return calls
}

// SetDisabled calls SetDisabledFunc.
func (mock *RegistryMock) SetDisabled(sensor string, state bool) error {
	if mock.SetDisabledFunc == nil {
		panic("RegistryMock.SetDisabledFunc: method is nil but Registry.SetDisabled was just called")
	}
	callInfo := struct {
		Sensor string
		State  bool
	}{
		Sensor: sensor,
		State:  state,
	}
	mock.lockSetDisabled.Lock()
	mock.calls.SetDisabled = append(mock.calls.SetDisabled, callInfo)
	mock.lockSetDisabled.Unlock()
	return mock.SetDisabledFunc(sensor, state)
}

// SetDisabledCalls gets all the calls that were made to SetDisabled.
// Check the length with:
//
//	len(mockedRegistry.SetDisabledCalls())
func (mock *RegistryMock) SetDisabledCalls() []struct {
	Sensor string
	State  bool
} {
	var calls []struct {
		Sensor string
		State  bool
	}
	mock.lockSetDisabled.RLock()
	calls = mock.calls.SetDisabled
	mock.lockSetDisabled.RUnlock()
	return calls
}

// SetRegistered calls SetRegisteredFunc.
func (mock *RegistryMock) SetRegistered(sensor string, state bool) error {
	if mock.SetRegisteredFunc == nil {
		panic("RegistryMock.SetRegisteredFunc: method is nil but Registry.SetRegistered was just called")
	}
	callInfo := struct {
		Sensor string
		State  bool
	}{
		Sensor: sensor,
		State:  state,
	}
	mock.lockSetRegistered.Lock()
	mock.calls.SetRegistered = append(mock.calls.SetRegistered, callInfo)
	mock.lockSetRegistered.Unlock()
	return mock.SetRegisteredFunc(sensor, state)
}

// SetRegisteredCalls gets all the calls that were made to SetRegistered.
// Check the length with:
//
//	len(mockedRegistry.SetRegisteredCalls())
func (mock *RegistryMock) SetRegisteredCalls() []struct {
	Sensor string
	State  bool
} {
	var calls []struct {
		Sensor string
		State  bool
	}
	mock.lockSetRegistered.RLock()
	calls = mock.calls.SetRegistered
	mock.lockSetRegistered.RUnlock()
	return calls
}
