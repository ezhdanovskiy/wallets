package http

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/ezhdanovskiy/wallets/internal/dto"
	"github.com/ezhdanovskiy/wallets/internal/http/mocks"
	"go.uber.org/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestServer_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	server := NewServer(zap.NewNop().Sugar(), 0, mockService)

	// Start server in goroutine
	go func() {
		_ = server.Run()
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Shutdown server
	server.Shutdown()
}

func TestServer_GetV1ApiRouters(t *testing.T) {
	server := &Server{
		log: zap.NewNop().Sugar(),
	}

	router := server.GetV1ApiRouters()
	assert.NotNil(t, router)
}

func TestServer_handlers_with_CSV_error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	server := &Server{
		log: zap.NewNop().Sugar(),
		svc: mockService,
	}

	// Test CSV conversion error
	req, _ := http.NewRequest("GET", "/v1/wallets/operations?wallet=wallet1&format=csv", nil)
	w := &errResponseWriter{}

	mockService.EXPECT().GetOperations(gomock.Any()).Return([]dto.Operation{
		{
			Wallet:    "wallet1",
			Type:      "deposit",
			Amount:    dto.Amount(10000),
			Timestamp: time.Unix(1234567890, 0).UTC(),
		},
	}, nil)

	server.getOperations(w, req)
	assert.Equal(t, http.StatusOK, w.statusCode)
}

func TestServer_writeResponse_marshal_error(t *testing.T) {
	server := &Server{
		log: zap.NewNop().Sugar(),
	}

	// Create channel that will cause json.Marshal to fail
	ch := make(chan int)
	
	w := &errResponseWriter{}
	server.writeResponse(w, http.StatusOK, ch)
	
	assert.Equal(t, http.StatusOK, w.statusCode)
}

func TestServer_writeErrorResponse_marshal_error(t *testing.T) {
	server := &Server{
		log: zap.NewNop().Sugar(),
	}

	// Create custom error that will cause json.Marshal to fail
	customErr := &customError{}
	
	w := &errResponseWriter{}
	server.writeErrorResponse(w, customErr)
	
	assert.Equal(t, http.StatusInternalServerError, w.statusCode)
}

// Helper types for testing edge cases

type errResponseWriter struct {
	statusCode int
	written    []byte
}

func (w *errResponseWriter) Header() http.Header {
	return http.Header{}
}

func (w *errResponseWriter) Write(b []byte) (int, error) {
	w.written = append(w.written, b...)
	return len(b), nil
}

func (w *errResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

type customError struct{}

func (e *customError) Error() string {
	return "custom error"
}

func (e *customError) MarshalJSON() ([]byte, error) {
	return nil, errors.New("marshal error")
}