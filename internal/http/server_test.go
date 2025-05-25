package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ezhdanovskiy/wallets/internal/httperr"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewServer(t *testing.T) {
	logger := zap.NewNop().Sugar()
	port := 8080
	
	server := NewServer(logger, port, nil)
	
	assert.NotNil(t, server)
	assert.Equal(t, logger, server.log)
	assert.Equal(t, port, server.httpPort)
	assert.Nil(t, server.svc)
}

func TestServer_writeResponse(t *testing.T) {
	server := &Server{
		log: zap.NewNop().Sugar(),
	}

	tests := []struct {
		name         string
		code         int
		payload      interface{}
		expectedBody string
		expectedCode int
	}{
		{
			name:         "nil payload",
			code:         http.StatusOK,
			payload:      nil,
			expectedBody: "",
			expectedCode: http.StatusOK,
		},
		{
			name:         "struct payload",
			code:         http.StatusOK,
			payload:      map[string]string{"key": "value"},
			expectedBody: `{"data":{"key":"value"}}`,
			expectedCode: http.StatusOK,
		},
		{
			name:         "byte array payload",
			code:         http.StatusOK,
			payload:      []byte(`raw data`),
			expectedBody: `raw data`,
			expectedCode: http.StatusOK,
		},
		{
			name:         "custom status code",
			code:         http.StatusCreated,
			payload:      map[string]int{"count": 42},
			expectedBody: `{"data":{"count":42}}`,
			expectedCode: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			
			server.writeResponse(rec, tt.code, tt.payload)
			
			assert.Equal(t, tt.expectedCode, rec.Code)
			assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("content-type"))
			
			if tt.expectedBody != "" {
				if _, ok := tt.payload.([]byte); ok {
					assert.Equal(t, tt.expectedBody, rec.Body.String())
				} else {
					assert.JSONEq(t, tt.expectedBody, rec.Body.String())
				}
			}
		})
	}
}

func TestServer_writeErrorResponse(t *testing.T) {
	server := &Server{
		log: zap.NewNop().Sugar(),
	}

	tests := []struct {
		name         string
		err          error
		expectedBody string
		expectedCode int
	}{
		{
			name:         "nil error",
			err:          nil,
			expectedBody: "",
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:         "regular error",
			err:          errors.New("something went wrong"),
			expectedBody: `{"error":"something went wrong"}`,
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:         "http error with custom status",
			err:          httperr.New(http.StatusBadRequest, "invalid request"),
			expectedBody: `{"error":"invalid request"}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "http error not found",
			err:          httperr.New(http.StatusNotFound, "resource not found"),
			expectedBody: `{"error":"resource not found"}`,
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			
			server.writeErrorResponse(rec, tt.err)
			
			assert.Equal(t, tt.expectedCode, rec.Code)
			assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("content-type"))
			
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, rec.Body.String())
			}
		})
	}
}

func TestResp_JSON(t *testing.T) {
	tests := []struct {
		name     string
		resp     Resp
		expected string
	}{
		{
			name:     "with error",
			resp:     Resp{Error: "error message"},
			expected: `{"error":"error message"}`,
		},
		{
			name:     "with data",
			resp:     Resp{Data: "some data"},
			expected: `{"data":"some data"}`,
		},
		{
			name:     "empty",
			resp:     Resp{},
			expected: `{}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.resp)
			assert.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(data))
		})
	}
}