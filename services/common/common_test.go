package common

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/google/uuid"

	"github.com/Dimitriy14/image-resizing/logger"
)

func init() {
	logger.Log = logger.NewMokLogger()
}

func Test_render(t *testing.T) {
	type Test struct {
		StatusCode  int
		BodyMessage string
	}
	var (
		getTestData = func(status int, msg string) Test {
			return Test{
				StatusCode:  status,
				BodyMessage: msg,
			}
		}

		tests = []Test{
			getTestData(http.StatusOK, `test 200`),
			getTestData(http.StatusCreated, `test 201`),
			getTestData(http.StatusNotFound, `test 404`),
			getTestData(http.StatusConflict, `test 409`),
		}
	)

	for i, test := range tests {
		t.Run(fmt.Sprintf(`Test case #%d`, i), func(t *testing.T) {
			var rr = httptest.NewRecorder()
			render(rr, test.StatusCode, []byte(test.BodyMessage))

			headers := rr.Header()
			if got := headers.Get("Content-Type"); got != "application/json; charset=utf-8" {
				t.Errorf(`invalid Content-Type header: %s`, got)
			}

			if got := headers.Get("Access-Control-Allow-Origin"); got != "*" {
				t.Errorf(`invalid Access-Control-Allow-Origin header: %s`, got)
			}

			if got := rr.Code; got != test.StatusCode {
				t.Errorf(`wrong StatusCode: want: %d, got: %d`, test.StatusCode, got)
			}

			if got := rr.Body.String(); got != test.BodyMessage {
				t.Errorf(`wrong Body: want: %s, got: %s`, test.BodyMessage, got)
			}
		})
	}
}

func Test_ReadRequestJSONBodyToStruct(t *testing.T) {
	type Test struct {
		Body              string
		DestinationStruct interface{}
		ExpectError       bool
	}

	var (
		getRequest = func(body string) *http.Request {
			return httptest.NewRequest(http.MethodPost, "http://localhost", bytes.NewReader([]byte(body)))
		}

		tests = []Test{
			{
				Body:              ``,
				DestinationStruct: struct{}{},
				ExpectError:       true,
			},
			{
				Body: `{"A":"str","B":42}`,
				DestinationStruct: struct {
					A string
					B int
				}{},
				ExpectError: false,
			},
			{
				Body: `{A":"tr""B":42`,
				DestinationStruct: struct {
					A string
					B int
				}{},
				ExpectError: true,
			},
		}
	)

	for i, test := range tests {
		t.Run(fmt.Sprintf(`Test case #%d`, i), func(t *testing.T) {
			err := ReadRequestJSONBodyToStruct(getRequest(test.Body), &test.DestinationStruct)
			if (err != nil) != test.ExpectError {
				t.Errorf(`got err: %v, but want err: %v`, err, test.ExpectError)
			}
		})
	}

}

func Test_SendConflictError(t *testing.T) {
	tests := []string{
		"good message",
		"yet another good message",
		"test",
	}

	for i, msg := range tests {
		t.Run(fmt.Sprintf(`Test case #%d`, i), func(t *testing.T) {
			var (
				rr     = httptest.NewRecorder()
				dec    = json.NewDecoder(rr.Body)
				errMsg ErrorMessage
			)

			SendConflictError(rr, msg)

			if got := rr.Code; got != http.StatusConflict {
				t.Errorf(`Status Code: got %d, want %d`, got, http.StatusConflict)
			}

			dec.Decode(&errMsg)
			if got := errMsg.Error.Message; got != msg {
				t.Errorf(`Body Message: got %s, want %s`, got, msg)
			}
		})
	}
}

func Test_RenderJSONCreated(t *testing.T) {
	type Test struct {
		Body           interface{}
		WantStatusCode int
	}

	var (
		getTestData = func(body interface{}, status int) Test {
			return Test{
				Body:           body,
				WantStatusCode: status,
			}
		}

		invalidBody = make(chan int)

		tests = []Test{
			getTestData(`good test`, http.StatusCreated),
			getTestData(struct{ A string }{A: `test`}, http.StatusCreated),
			getTestData(invalidBody, http.StatusInternalServerError),
		}
	)

	for i, test := range tests {
		t.Run(fmt.Sprintf(`Test case #%d`, i), func(t *testing.T) {
			var rr = httptest.NewRecorder()
			RenderJSONCreated(rr, test.Body)
			if got := rr.Code; got != test.WantStatusCode {
				t.Errorf(`Status Code: got %v, want %v`, got, test.WantStatusCode)
			}
		})
	}
}

func Test_RenderJSON(t *testing.T) {
	type Test struct {
		Body           interface{}
		WantStatusCode int
	}

	var (
		getTestData = func(body interface{}, status int) Test {
			return Test{
				Body:           body,
				WantStatusCode: status,
			}
		}

		invalidBody = make(chan int)

		tests = []Test{
			getTestData(`good test`, http.StatusOK),
			getTestData(struct{ A string }{A: `test`}, http.StatusOK),
			getTestData(invalidBody, http.StatusInternalServerError),
		}
	)

	for i, test := range tests {
		t.Run(fmt.Sprintf(`Test case #%d`, i), func(t *testing.T) {
			var rr = httptest.NewRecorder()
			RenderJSON(rr, test.Body)
			if got := rr.Code; got != test.WantStatusCode {
				t.Errorf(`Status Code: got %v, want %v`, got, test.WantStatusCode)
			}
		})
	}
}

func TestSendNotFound(t *testing.T) {
	testCases := []struct {
		format string
		args   []interface{}
		resp   string
	}{
		{"msg: %s", []interface{}{"empty"}, `{"error":{"message":"msg: empty"}}`},
		{"text", nil, `{"error":{"message":"text"}}`},
	}

	for _, tc := range testCases {
		w := httptest.NewRecorder()
		SendNotFound(w, tc.format, tc.args...)
		if w.Code != http.StatusNotFound {
			t.Errorf("want code %d but got %d", http.StatusNotFound, w.Code)
		}
		if w.Body.String() != tc.resp {
			t.Errorf("want body [%s] but got [%s]", tc.resp, w.Body)
		}
	}
}

func Test_CloseWithErrCheck(t *testing.T) {

	tests := []struct {
		name             string
		wantNativeLogger bool
		wantErr          bool
	}{
		{
			name:             `bad`,
			wantNativeLogger: true,
			wantErr:          true,
		},
		{
			name:             `bad - 2`,
			wantNativeLogger: false,
			wantErr:          true,
		},
		{
			name:    `good`,
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var (
				fc = &FakeCloser{
					NeedError: test.wantErr,
				}
			)

			CloseWithErrCheck(fc, test.name)

			if test.wantErr == fc.Closed {
				t.Error(`Closer hadn't been closed`)
			}
		})
	}
}

type FakeCloser struct {
	NeedError bool
	Closed    bool
}

func (fc *FakeCloser) Close() error {
	if fc.NeedError {
		return errors.New(`closer failed`)
	}
	fc.Closed = true
	return nil
}

func TestGetUserIDFromCtx(t *testing.T) {
	var (
		nonEmptyUUID = uuid.New()
	)

	ctx := context.WithValue(context.Background(), UserID, nonEmptyUUID)

	assert.EqualValues(t, nonEmptyUUID, GetUserIDFromCtx(ctx))
	assert.EqualValues(t, uuid.UUID{}, GetUserIDFromCtx(context.Background()))
}
