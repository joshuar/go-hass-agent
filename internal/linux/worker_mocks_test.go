// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package linux

import (
	"context"
	"github.com/joshuar/go-hass-agent/internal/models"
	"sync"
	"time"
)

// Ensure, that PollingSensorTypeMock does implement PollingSensorType.
// If this is not the case, regenerate this file with moq.
var _ PollingSensorType = &PollingSensorTypeMock{}

// PollingSensorTypeMock is a mock implementation of PollingSensorType.
//
//	func TestSomethingThatUsesPollingSensorType(t *testing.T) {
//
//		// make and configure a mocked PollingSensorType
//		mockedPollingSensorType := &PollingSensorTypeMock{
//			SensorsFunc: func(ctx context.Context) ([]models.Entity, error) {
//				panic("mock out the Sensors method")
//			},
//			UpdateDeltaFunc: func(delta time.Duration)  {
//				panic("mock out the UpdateDelta method")
//			},
//		}
//
//		// use mockedPollingSensorType in code that requires PollingSensorType
//		// and then make assertions.
//
//	}
type PollingSensorTypeMock struct {
	// SensorsFunc mocks the Sensors method.
	SensorsFunc func(ctx context.Context) ([]models.Entity, error)

	// UpdateDeltaFunc mocks the UpdateDelta method.
	UpdateDeltaFunc func(delta time.Duration)

	// calls tracks calls to the methods.
	calls struct {
		// Sensors holds details about calls to the Sensors method.
		Sensors []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
		// UpdateDelta holds details about calls to the UpdateDelta method.
		UpdateDelta []struct {
			// Delta is the delta argument value.
			Delta time.Duration
		}
	}
	lockSensors     sync.RWMutex
	lockUpdateDelta sync.RWMutex
}

// Sensors calls SensorsFunc.
func (mock *PollingSensorTypeMock) Sensors(ctx context.Context) ([]models.Entity, error) {
	if mock.SensorsFunc == nil {
		panic("PollingSensorTypeMock.SensorsFunc: method is nil but PollingSensorType.Sensors was just called")
	}
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	mock.lockSensors.Lock()
	mock.calls.Sensors = append(mock.calls.Sensors, callInfo)
	mock.lockSensors.Unlock()
	return mock.SensorsFunc(ctx)
}

// SensorsCalls gets all the calls that were made to Sensors.
// Check the length with:
//
//	len(mockedPollingSensorType.SensorsCalls())
func (mock *PollingSensorTypeMock) SensorsCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	mock.lockSensors.RLock()
	calls = mock.calls.Sensors
	mock.lockSensors.RUnlock()
	return calls
}

// UpdateDelta calls UpdateDeltaFunc.
func (mock *PollingSensorTypeMock) UpdateDelta(delta time.Duration) {
	if mock.UpdateDeltaFunc == nil {
		panic("PollingSensorTypeMock.UpdateDeltaFunc: method is nil but PollingSensorType.UpdateDelta was just called")
	}
	callInfo := struct {
		Delta time.Duration
	}{
		Delta: delta,
	}
	mock.lockUpdateDelta.Lock()
	mock.calls.UpdateDelta = append(mock.calls.UpdateDelta, callInfo)
	mock.lockUpdateDelta.Unlock()
	mock.UpdateDeltaFunc(delta)
}

// UpdateDeltaCalls gets all the calls that were made to UpdateDelta.
// Check the length with:
//
//	len(mockedPollingSensorType.UpdateDeltaCalls())
func (mock *PollingSensorTypeMock) UpdateDeltaCalls() []struct {
	Delta time.Duration
} {
	var calls []struct {
		Delta time.Duration
	}
	mock.lockUpdateDelta.RLock()
	calls = mock.calls.UpdateDelta
	mock.lockUpdateDelta.RUnlock()
	return calls
}

// Ensure, that EventSensorTypeMock does implement EventSensorType.
// If this is not the case, regenerate this file with moq.
var _ EventSensorType = &EventSensorTypeMock{}

// EventSensorTypeMock is a mock implementation of EventSensorType.
//
//	func TestSomethingThatUsesEventSensorType(t *testing.T) {
//
//		// make and configure a mocked EventSensorType
//		mockedEventSensorType := &EventSensorTypeMock{
//			EventsFunc: func(ctx context.Context) (<-chan models.Entity, error) {
//				panic("mock out the Events method")
//			},
//			SensorsFunc: func(ctx context.Context) ([]models.Entity, error) {
//				panic("mock out the Sensors method")
//			},
//		}
//
//		// use mockedEventSensorType in code that requires EventSensorType
//		// and then make assertions.
//
//	}
type EventSensorTypeMock struct {
	// EventsFunc mocks the Events method.
	EventsFunc func(ctx context.Context) (<-chan models.Entity, error)

	// SensorsFunc mocks the Sensors method.
	SensorsFunc func(ctx context.Context) ([]models.Entity, error)

	// calls tracks calls to the methods.
	calls struct {
		// Events holds details about calls to the Events method.
		Events []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
		// Sensors holds details about calls to the Sensors method.
		Sensors []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
	}
	lockEvents  sync.RWMutex
	lockSensors sync.RWMutex
}

// Events calls EventsFunc.
func (mock *EventSensorTypeMock) Events(ctx context.Context) (<-chan models.Entity, error) {
	if mock.EventsFunc == nil {
		panic("EventSensorTypeMock.EventsFunc: method is nil but EventSensorType.Events was just called")
	}
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	mock.lockEvents.Lock()
	mock.calls.Events = append(mock.calls.Events, callInfo)
	mock.lockEvents.Unlock()
	return mock.EventsFunc(ctx)
}

// EventsCalls gets all the calls that were made to Events.
// Check the length with:
//
//	len(mockedEventSensorType.EventsCalls())
func (mock *EventSensorTypeMock) EventsCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	mock.lockEvents.RLock()
	calls = mock.calls.Events
	mock.lockEvents.RUnlock()
	return calls
}

// Sensors calls SensorsFunc.
func (mock *EventSensorTypeMock) Sensors(ctx context.Context) ([]models.Entity, error) {
	if mock.SensorsFunc == nil {
		panic("EventSensorTypeMock.SensorsFunc: method is nil but EventSensorType.Sensors was just called")
	}
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	mock.lockSensors.Lock()
	mock.calls.Sensors = append(mock.calls.Sensors, callInfo)
	mock.lockSensors.Unlock()
	return mock.SensorsFunc(ctx)
}

// SensorsCalls gets all the calls that were made to Sensors.
// Check the length with:
//
//	len(mockedEventSensorType.SensorsCalls())
func (mock *EventSensorTypeMock) SensorsCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	mock.lockSensors.RLock()
	calls = mock.calls.Sensors
	mock.lockSensors.RUnlock()
	return calls
}

// Ensure, that OneShotSensorTypeMock does implement OneShotSensorType.
// If this is not the case, regenerate this file with moq.
var _ OneShotSensorType = &OneShotSensorTypeMock{}

// OneShotSensorTypeMock is a mock implementation of OneShotSensorType.
//
//	func TestSomethingThatUsesOneShotSensorType(t *testing.T) {
//
//		// make and configure a mocked OneShotSensorType
//		mockedOneShotSensorType := &OneShotSensorTypeMock{
//			SensorsFunc: func(ctx context.Context) ([]models.Entity, error) {
//				panic("mock out the Sensors method")
//			},
//		}
//
//		// use mockedOneShotSensorType in code that requires OneShotSensorType
//		// and then make assertions.
//
//	}
type OneShotSensorTypeMock struct {
	// SensorsFunc mocks the Sensors method.
	SensorsFunc func(ctx context.Context) ([]models.Entity, error)

	// calls tracks calls to the methods.
	calls struct {
		// Sensors holds details about calls to the Sensors method.
		Sensors []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
	}
	lockSensors sync.RWMutex
}

// Sensors calls SensorsFunc.
func (mock *OneShotSensorTypeMock) Sensors(ctx context.Context) ([]models.Entity, error) {
	if mock.SensorsFunc == nil {
		panic("OneShotSensorTypeMock.SensorsFunc: method is nil but OneShotSensorType.Sensors was just called")
	}
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	mock.lockSensors.Lock()
	mock.calls.Sensors = append(mock.calls.Sensors, callInfo)
	mock.lockSensors.Unlock()
	return mock.SensorsFunc(ctx)
}

// SensorsCalls gets all the calls that were made to Sensors.
// Check the length with:
//
//	len(mockedOneShotSensorType.SensorsCalls())
func (mock *OneShotSensorTypeMock) SensorsCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	mock.lockSensors.RLock()
	calls = mock.calls.Sensors
	mock.lockSensors.RUnlock()
	return calls
}

// Ensure, that EventTypeMock does implement EventType.
// If this is not the case, regenerate this file with moq.
var _ EventType = &EventTypeMock{}

// EventTypeMock is a mock implementation of EventType.
//
//	func TestSomethingThatUsesEventType(t *testing.T) {
//
//		// make and configure a mocked EventType
//		mockedEventType := &EventTypeMock{
//			EventsFunc: func(ctx context.Context) (<-chan models.Entity, error) {
//				panic("mock out the Events method")
//			},
//		}
//
//		// use mockedEventType in code that requires EventType
//		// and then make assertions.
//
//	}
type EventTypeMock struct {
	// EventsFunc mocks the Events method.
	EventsFunc func(ctx context.Context) (<-chan models.Entity, error)

	// calls tracks calls to the methods.
	calls struct {
		// Events holds details about calls to the Events method.
		Events []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
	}
	lockEvents sync.RWMutex
}

// Events calls EventsFunc.
func (mock *EventTypeMock) Events(ctx context.Context) (<-chan models.Entity, error) {
	if mock.EventsFunc == nil {
		panic("EventTypeMock.EventsFunc: method is nil but EventType.Events was just called")
	}
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	mock.lockEvents.Lock()
	mock.calls.Events = append(mock.calls.Events, callInfo)
	mock.lockEvents.Unlock()
	return mock.EventsFunc(ctx)
}

// EventsCalls gets all the calls that were made to Events.
// Check the length with:
//
//	len(mockedEventType.EventsCalls())
func (mock *EventTypeMock) EventsCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	mock.lockEvents.RLock()
	calls = mock.calls.Events
	mock.lockEvents.RUnlock()
	return calls
}
