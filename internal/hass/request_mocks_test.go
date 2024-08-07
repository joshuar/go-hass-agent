// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package hass

import (
	"encoding/json"
	"sync"
)

// Ensure, that PostRequestMock does implement PostRequest.
// If this is not the case, regenerate this file with moq.
var _ PostRequest = &PostRequestMock{}

// PostRequestMock is a mock implementation of PostRequest.
//
//	func TestSomethingThatUsesPostRequest(t *testing.T) {
//
//		// make and configure a mocked PostRequest
//		mockedPostRequest := &PostRequestMock{
//			RequestBodyFunc: func() json.RawMessage {
//				panic("mock out the RequestBody method")
//			},
//		}
//
//		// use mockedPostRequest in code that requires PostRequest
//		// and then make assertions.
//
//	}
type PostRequestMock struct {
	// RequestBodyFunc mocks the RequestBody method.
	RequestBodyFunc func() json.RawMessage

	// calls tracks calls to the methods.
	calls struct {
		// RequestBody holds details about calls to the RequestBody method.
		RequestBody []struct {
		}
	}
	lockRequestBody sync.RWMutex
}

// RequestBody calls RequestBodyFunc.
func (mock *PostRequestMock) RequestBody() json.RawMessage {
	if mock.RequestBodyFunc == nil {
		panic("PostRequestMock.RequestBodyFunc: method is nil but PostRequest.RequestBody was just called")
	}
	callInfo := struct {
	}{}
	mock.lockRequestBody.Lock()
	mock.calls.RequestBody = append(mock.calls.RequestBody, callInfo)
	mock.lockRequestBody.Unlock()
	return mock.RequestBodyFunc()
}

// RequestBodyCalls gets all the calls that were made to RequestBody.
// Check the length with:
//
//	len(mockedPostRequest.RequestBodyCalls())
func (mock *PostRequestMock) RequestBodyCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockRequestBody.RLock()
	calls = mock.calls.RequestBody
	mock.lockRequestBody.RUnlock()
	return calls
}

// Ensure, that ResponseMock does implement Response.
// If this is not the case, regenerate this file with moq.
var _ Response = &ResponseMock{}

// ResponseMock is a mock implementation of Response.
//
//	func TestSomethingThatUsesResponse(t *testing.T) {
//
//		// make and configure a mocked Response
//		mockedResponse := &ResponseMock{
//			ErrorFunc: func() string {
//				panic("mock out the Error method")
//			},
//			UnmarshalErrorFunc: func(data []byte) error {
//				panic("mock out the UnmarshalError method")
//			},
//			UnmarshalJSONFunc: func(bytes []byte) error {
//				panic("mock out the UnmarshalJSON method")
//			},
//		}
//
//		// use mockedResponse in code that requires Response
//		// and then make assertions.
//
//	}
type ResponseMock struct {
	// ErrorFunc mocks the Error method.
	ErrorFunc func() string

	// UnmarshalErrorFunc mocks the UnmarshalError method.
	UnmarshalErrorFunc func(data []byte) error

	// UnmarshalJSONFunc mocks the UnmarshalJSON method.
	UnmarshalJSONFunc func(bytes []byte) error

	// calls tracks calls to the methods.
	calls struct {
		// Error holds details about calls to the Error method.
		Error []struct {
		}
		// UnmarshalError holds details about calls to the UnmarshalError method.
		UnmarshalError []struct {
			// Data is the data argument value.
			Data []byte
		}
		// UnmarshalJSON holds details about calls to the UnmarshalJSON method.
		UnmarshalJSON []struct {
			// Bytes is the bytes argument value.
			Bytes []byte
		}
	}
	lockError          sync.RWMutex
	lockUnmarshalError sync.RWMutex
	lockUnmarshalJSON  sync.RWMutex
}

// Error calls ErrorFunc.
func (mock *ResponseMock) Error() string {
	if mock.ErrorFunc == nil {
		panic("ResponseMock.ErrorFunc: method is nil but Response.Error was just called")
	}
	callInfo := struct {
	}{}
	mock.lockError.Lock()
	mock.calls.Error = append(mock.calls.Error, callInfo)
	mock.lockError.Unlock()
	return mock.ErrorFunc()
}

// ErrorCalls gets all the calls that were made to Error.
// Check the length with:
//
//	len(mockedResponse.ErrorCalls())
func (mock *ResponseMock) ErrorCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockError.RLock()
	calls = mock.calls.Error
	mock.lockError.RUnlock()
	return calls
}

// UnmarshalError calls UnmarshalErrorFunc.
func (mock *ResponseMock) UnmarshalError(data []byte) error {
	if mock.UnmarshalErrorFunc == nil {
		panic("ResponseMock.UnmarshalErrorFunc: method is nil but Response.UnmarshalError was just called")
	}
	callInfo := struct {
		Data []byte
	}{
		Data: data,
	}
	mock.lockUnmarshalError.Lock()
	mock.calls.UnmarshalError = append(mock.calls.UnmarshalError, callInfo)
	mock.lockUnmarshalError.Unlock()
	return mock.UnmarshalErrorFunc(data)
}

// UnmarshalErrorCalls gets all the calls that were made to UnmarshalError.
// Check the length with:
//
//	len(mockedResponse.UnmarshalErrorCalls())
func (mock *ResponseMock) UnmarshalErrorCalls() []struct {
	Data []byte
} {
	var calls []struct {
		Data []byte
	}
	mock.lockUnmarshalError.RLock()
	calls = mock.calls.UnmarshalError
	mock.lockUnmarshalError.RUnlock()
	return calls
}

// UnmarshalJSON calls UnmarshalJSONFunc.
func (mock *ResponseMock) UnmarshalJSON(bytes []byte) error {
	if mock.UnmarshalJSONFunc == nil {
		panic("ResponseMock.UnmarshalJSONFunc: method is nil but Response.UnmarshalJSON was just called")
	}
	callInfo := struct {
		Bytes []byte
	}{
		Bytes: bytes,
	}
	mock.lockUnmarshalJSON.Lock()
	mock.calls.UnmarshalJSON = append(mock.calls.UnmarshalJSON, callInfo)
	mock.lockUnmarshalJSON.Unlock()
	return mock.UnmarshalJSONFunc(bytes)
}

// UnmarshalJSONCalls gets all the calls that were made to UnmarshalJSON.
// Check the length with:
//
//	len(mockedResponse.UnmarshalJSONCalls())
func (mock *ResponseMock) UnmarshalJSONCalls() []struct {
	Bytes []byte
} {
	var calls []struct {
		Bytes []byte
	}
	mock.lockUnmarshalJSON.RLock()
	calls = mock.calls.UnmarshalJSON
	mock.lockUnmarshalJSON.RUnlock()
	return calls
}
