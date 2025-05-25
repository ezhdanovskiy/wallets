package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ezhdanovskiy/wallets/internal/dto"
	"github.com/ezhdanovskiy/wallets/internal/http/mocks"
	"github.com/ezhdanovskiy/wallets/internal/httperr"
	"go.uber.org/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestServer_createWallet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	server := &Server{
		log: zap.NewNop().Sugar(),
		svc: mockService,
	}

	tests := []struct {
		name           string
		body           interface{}
		mockSetup      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success",
			body: dto.CreateWalletRequest{
				Name: "Test Wallet",
			},
			mockSetup: func() {
				mockService.EXPECT().CreateWallet(dto.CreateWalletRequest{
					Name: "Test Wallet",
				}).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:           "invalid json",
			body:           "invalid json",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"failed to decode body"}`,
		},
		{
			name: "service error",
			body: dto.CreateWalletRequest{
				Name: "Test Wallet",
			},
			mockSetup: func() {
				mockService.EXPECT().CreateWallet(gomock.Any()).Return(errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"service error"}`,
		},
		{
			name: "http error from service",
			body: dto.CreateWalletRequest{
				Name: "Test Wallet",
			},
			mockSetup: func() {
				mockService.EXPECT().CreateWallet(gomock.Any()).Return(httperr.New(http.StatusConflict, "wallet already exists"))
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   `{"error":"wallet already exists"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			var body []byte
			if str, ok := tt.body.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/v1/wallets", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()

			server.createWallet(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, rec.Body.String())
			}
		})
	}
}

func TestServer_deposit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	server := &Server{
		log: zap.NewNop().Sugar(),
		svc: mockService,
	}

	tests := []struct {
		name           string
		body           interface{}
		mockSetup      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success",
			body: dto.Deposit{
				Wallet: "wallet1",
				Amount: dto.Amount(10050),
			},
			mockSetup: func() {
				mockService.EXPECT().IncreaseWalletBalance(dto.Deposit{
					Wallet: "wallet1",
					Amount: dto.Amount(10050),
				}).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:           "invalid json",
			body:           "invalid json",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"failed to decode body"}`,
		},
		{
			name: "service error",
			body: dto.Deposit{
				Wallet: "wallet1",
				Amount: dto.Amount(10050),
			},
			mockSetup: func() {
				mockService.EXPECT().IncreaseWalletBalance(gomock.Any()).Return(errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"service error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			var body []byte
			if str, ok := tt.body.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/v1/wallets/deposit", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()

			server.deposit(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, rec.Body.String())
			}
		})
	}
}

func TestServer_transfer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	server := &Server{
		log: zap.NewNop().Sugar(),
		svc: mockService,
	}

	tests := []struct {
		name           string
		body           interface{}
		mockSetup      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success",
			body: dto.Transfer{
				WalletFrom: "wallet1",
				WalletTo:   "wallet2",
				Amount:     dto.Amount(5000),
			},
			mockSetup: func() {
				mockService.EXPECT().Transfer(dto.Transfer{
					WalletFrom: "wallet1",
					WalletTo:   "wallet2",
					Amount:     dto.Amount(5000),
				}).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:           "invalid json",
			body:           "invalid json",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"failed to decode body"}`,
		},
		{
			name: "service error",
			body: dto.Transfer{
				WalletFrom: "wallet1",
				WalletTo:   "wallet2",
				Amount:     dto.Amount(5000),
			},
			mockSetup: func() {
				mockService.EXPECT().Transfer(gomock.Any()).Return(errors.New("insufficient funds"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"insufficient funds"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			var body []byte
			if str, ok := tt.body.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/v1/wallets/transfer", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()

			server.transfer(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, rec.Body.String())
			}
		})
	}
}

func TestServer_getOperations(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	server := &Server{
		log: zap.NewNop().Sugar(),
		svc: mockService,
	}

	tests := []struct {
		name           string
		url            string
		mockSetup      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success with default limit",
			url:  "/v1/wallets/operations?wallet=wallet1",
			mockSetup: func() {
				mockService.EXPECT().GetOperations(dto.OperationsFilter{
					Wallet: "wallet1",
					Limit:  20,
				}).Return([]dto.Operation{
					{
						Wallet:    "wallet1",
						Type:      "deposit",
						Amount:    dto.Amount(10000),
						Timestamp: time.Unix(1234567890, 0).UTC(),
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: `{"data":[{
				"wallet":"wallet1",
				"type":"deposit",
				"amount":10000,
				"other_wallet":"",
				"timestamp":"2009-02-13T23:31:30Z"
			}]}`,
		},
		{
			name:           "missing wallet parameter",
			url:            "/v1/wallets/operations",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"empty wallet parameter"}`,
		},
		{
			name:           "invalid start_date",
			url:            "/v1/wallets/operations?wallet=wallet1&start_date=invalid",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"failed to parse start_date"}`,
		},
		{
			name:           "invalid end_date",
			url:            "/v1/wallets/operations?wallet=wallet1&end_date=invalid",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"failed to parse end_date"}`,
		},
		{
			name:           "invalid limit",
			url:            "/v1/wallets/operations?wallet=wallet1&limit=invalid",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"failed to parse limit"}`,
		},
		{
			name:           "limit out of range",
			url:            "/v1/wallets/operations?wallet=wallet1&limit=2000",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"wrong limit, it have to be in [1, 100]"}`,
		},
		{
			name:           "invalid offset",
			url:            "/v1/wallets/operations?wallet=wallet1&offset=invalid",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"failed to parse offset"}`,
		},
		{
			name: "with all parameters",
			url:  "/v1/wallets/operations?wallet=wallet1&type=deposit&start_date=1234567890&end_date=1234567899&limit=50&offset=10",
			mockSetup: func() {
				mockService.EXPECT().GetOperations(dto.OperationsFilter{
					Wallet:    "wallet1",
					Type:      "deposit",
					StartDate: 1234567890,
					EndDate:   1234567899,
					Limit:     50,
					Offset:    10,
				}).Return([]dto.Operation{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"data":[]}`,
		},
		{
			name: "service error",
			url:  "/v1/wallets/operations?wallet=wallet1",
			mockSetup: func() {
				mockService.EXPECT().GetOperations(gomock.Any()).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"database error"}`,
		},
		{
			name: "csv format",
			url:  "/v1/wallets/operations?wallet=wallet1&format=csv",
			mockSetup: func() {
				mockService.EXPECT().GetOperations(dto.OperationsFilter{
					Wallet: "wallet1",
					Limit:  20,
				}).Return([]dto.Operation{
					{
						Wallet:    "wallet1",
						Type:      "deposit",
						Amount:    dto.Amount(10000),
						Timestamp: time.Unix(1234567890, 0).UTC(),
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			rec := httptest.NewRecorder()

			server.getOperations(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, rec.Body.String())
			}
		})
	}
}