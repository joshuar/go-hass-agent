// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package agent

import (
	"context"
	"github.com/joshuar/go-hass-agent/internal/agent/ui"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"sync"
)

// Ensure, that UIMock does implement UI.
// If this is not the case, regenerate this file with moq.
var _ UI = &UIMock{}

// UIMock is a mock implementation of UI.
//
//	func TestSomethingThatUsesUI(t *testing.T) {
//
//		// make and configure a mocked UI
//		mockedUI := &UIMock{
//			DisplayNotificationFunc: func(n ui.Notification)  {
//				panic("mock out the DisplayNotification method")
//			},
//			DisplayRegistrationWindowFunc: func(ctx context.Context, input *hass.RegistrationInput, doneCh chan struct{})  {
//				panic("mock out the DisplayRegistrationWindow method")
//			},
//			DisplayTrayIconFunc: func(agent ui.Agent, trk ui.SensorTracker)  {
//				panic("mock out the DisplayTrayIcon method")
//			},
//			RunFunc: func(doneCh chan struct{})  {
//				panic("mock out the Run method")
//			},
//		}
//
//		// use mockedUI in code that requires UI
//		// and then make assertions.
//
//	}
type UIMock struct {
	// DisplayNotificationFunc mocks the DisplayNotification method.
	DisplayNotificationFunc func(n ui.Notification)

	// DisplayRegistrationWindowFunc mocks the DisplayRegistrationWindow method.
	DisplayRegistrationWindowFunc func(ctx context.Context, input *hass.RegistrationInput, doneCh chan struct{})

	// DisplayTrayIconFunc mocks the DisplayTrayIcon method.
	DisplayTrayIconFunc func(agent ui.Agent, trk ui.SensorTracker)

	// RunFunc mocks the Run method.
	RunFunc func(doneCh chan struct{})

	// calls tracks calls to the methods.
	calls struct {
		// DisplayNotification holds details about calls to the DisplayNotification method.
		DisplayNotification []struct {
			// N is the n argument value.
			N ui.Notification
		}
		// DisplayRegistrationWindow holds details about calls to the DisplayRegistrationWindow method.
		DisplayRegistrationWindow []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Input is the input argument value.
			Input *hass.RegistrationInput
			// DoneCh is the doneCh argument value.
			DoneCh chan struct{}
		}
		// DisplayTrayIcon holds details about calls to the DisplayTrayIcon method.
		DisplayTrayIcon []struct {
			// Agent is the agent argument value.
			Agent ui.Agent
			// Trk is the trk argument value.
			Trk ui.SensorTracker
		}
		// Run holds details about calls to the Run method.
		Run []struct {
			// DoneCh is the doneCh argument value.
			DoneCh chan struct{}
		}
	}
	lockDisplayNotification       sync.RWMutex
	lockDisplayRegistrationWindow sync.RWMutex
	lockDisplayTrayIcon           sync.RWMutex
	lockRun                       sync.RWMutex
}

// DisplayNotification calls DisplayNotificationFunc.
func (mock *UIMock) DisplayNotification(n ui.Notification) {
	if mock.DisplayNotificationFunc == nil {
		panic("UIMock.DisplayNotificationFunc: method is nil but UI.DisplayNotification was just called")
	}
	callInfo := struct {
		N ui.Notification
	}{
		N: n,
	}
	mock.lockDisplayNotification.Lock()
	mock.calls.DisplayNotification = append(mock.calls.DisplayNotification, callInfo)
	mock.lockDisplayNotification.Unlock()
	mock.DisplayNotificationFunc(n)
}

// DisplayNotificationCalls gets all the calls that were made to DisplayNotification.
// Check the length with:
//
//	len(mockedUI.DisplayNotificationCalls())
func (mock *UIMock) DisplayNotificationCalls() []struct {
	N ui.Notification
} {
	var calls []struct {
		N ui.Notification
	}
	mock.lockDisplayNotification.RLock()
	calls = mock.calls.DisplayNotification
	mock.lockDisplayNotification.RUnlock()
	return calls
}

// DisplayRegistrationWindow calls DisplayRegistrationWindowFunc.
func (mock *UIMock) DisplayRegistrationWindow(ctx context.Context, input *hass.RegistrationInput, doneCh chan struct{}) {
	if mock.DisplayRegistrationWindowFunc == nil {
		panic("UIMock.DisplayRegistrationWindowFunc: method is nil but UI.DisplayRegistrationWindow was just called")
	}
	callInfo := struct {
		Ctx    context.Context
		Input  *hass.RegistrationInput
		DoneCh chan struct{}
	}{
		Ctx:    ctx,
		Input:  input,
		DoneCh: doneCh,
	}
	mock.lockDisplayRegistrationWindow.Lock()
	mock.calls.DisplayRegistrationWindow = append(mock.calls.DisplayRegistrationWindow, callInfo)
	mock.lockDisplayRegistrationWindow.Unlock()
	mock.DisplayRegistrationWindowFunc(ctx, input, doneCh)
}

// DisplayRegistrationWindowCalls gets all the calls that were made to DisplayRegistrationWindow.
// Check the length with:
//
//	len(mockedUI.DisplayRegistrationWindowCalls())
func (mock *UIMock) DisplayRegistrationWindowCalls() []struct {
	Ctx    context.Context
	Input  *hass.RegistrationInput
	DoneCh chan struct{}
} {
	var calls []struct {
		Ctx    context.Context
		Input  *hass.RegistrationInput
		DoneCh chan struct{}
	}
	mock.lockDisplayRegistrationWindow.RLock()
	calls = mock.calls.DisplayRegistrationWindow
	mock.lockDisplayRegistrationWindow.RUnlock()
	return calls
}

// DisplayTrayIcon calls DisplayTrayIconFunc.
func (mock *UIMock) DisplayTrayIcon(agent ui.Agent, trk ui.SensorTracker) {
	if mock.DisplayTrayIconFunc == nil {
		panic("UIMock.DisplayTrayIconFunc: method is nil but UI.DisplayTrayIcon was just called")
	}
	callInfo := struct {
		Agent ui.Agent
		Trk   ui.SensorTracker
	}{
		Agent: agent,
		Trk:   trk,
	}
	mock.lockDisplayTrayIcon.Lock()
	mock.calls.DisplayTrayIcon = append(mock.calls.DisplayTrayIcon, callInfo)
	mock.lockDisplayTrayIcon.Unlock()
	mock.DisplayTrayIconFunc(agent, trk)
}

// DisplayTrayIconCalls gets all the calls that were made to DisplayTrayIcon.
// Check the length with:
//
//	len(mockedUI.DisplayTrayIconCalls())
func (mock *UIMock) DisplayTrayIconCalls() []struct {
	Agent ui.Agent
	Trk   ui.SensorTracker
} {
	var calls []struct {
		Agent ui.Agent
		Trk   ui.SensorTracker
	}
	mock.lockDisplayTrayIcon.RLock()
	calls = mock.calls.DisplayTrayIcon
	mock.lockDisplayTrayIcon.RUnlock()
	return calls
}

// Run calls RunFunc.
func (mock *UIMock) Run(doneCh chan struct{}) {
	if mock.RunFunc == nil {
		panic("UIMock.RunFunc: method is nil but UI.Run was just called")
	}
	callInfo := struct {
		DoneCh chan struct{}
	}{
		DoneCh: doneCh,
	}
	mock.lockRun.Lock()
	mock.calls.Run = append(mock.calls.Run, callInfo)
	mock.lockRun.Unlock()
	mock.RunFunc(doneCh)
}

// RunCalls gets all the calls that were made to Run.
// Check the length with:
//
//	len(mockedUI.RunCalls())
func (mock *UIMock) RunCalls() []struct {
	DoneCh chan struct{}
} {
	var calls []struct {
		DoneCh chan struct{}
	}
	mock.lockRun.RLock()
	calls = mock.calls.Run
	mock.lockRun.RUnlock()
	return calls
}
