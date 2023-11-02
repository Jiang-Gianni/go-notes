package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHelloWorldHandler(t *testing.T) {
	// TODO test cases
	tests := []struct {
		name           string
		in             *http.Request
		out            *httptest.ResponseRecorder
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Hello World",
			in:             httptest.NewRequest("GET", "/", nil),
			out:            httptest.NewRecorder(),
			expectedStatus: http.StatusOK,
			expectedBody:   "Hello, World!",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			helloWorldHandler(test.out, test.in)
			if test.out.Code != test.expectedStatus {
				t.Logf("Status expected: %d\n got: %d\n", test.expectedStatus, test.out.Code)
				t.Fail()
			}
			body := test.out.Body.String()
			if body != test.expectedBody {
				t.Logf("Body expected: %s\n got: %s\n", test.expectedBody, body)
				t.Fail()
			}
		})
	}

}
