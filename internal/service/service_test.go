package service

import (
	"database/sql"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/ezhdanovskiy/wallets/internal/dto"
	"github.com/ezhdanovskiy/wallets/internal/httperr"
	"github.com/ezhdanovskiy/wallets/internal/service/mocks"
)

const logsEnabled = false

const (
	testWalletName            = "WalletName01"
	testAmount     dto.Amount = 123.45
)

func TestService_CreateWallet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ts := newTestService(t)
		defer ts.Finish()

		ts.mockRepo.EXPECT().CreateWallet(testWalletName).
			Return(nil)

		req := dto.CreateWalletRequest{Name: testWalletName}
		err := ts.svc.CreateWallet(req)
		require.NoError(t, err)
	})

	t.Run("database connection error", func(t *testing.T) {
		ts := newTestService(t)
		defer ts.Finish()

		ts.mockRepo.EXPECT().CreateWallet(testWalletName).
			Return(sql.ErrConnDone)

		req := dto.CreateWalletRequest{Name: testWalletName}
		err := ts.svc.CreateWallet(req)
		require.Error(t, err)
		e, ok := err.(*httperr.Error)
		require.True(t, ok)
		assert.Equal(t, sql.ErrConnDone, e.Err)
		assert.Equal(t, ErrDatabase.StatusCode, e.StatusCode)
		assert.Equal(t, ErrDatabase.Message, e.Message)
	})
}

func TestService_IncreaseWalletBalance(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ts := newTestService(t)
		defer ts.Finish()

		ts.mockRepo.EXPECT().GetWallet(testWalletName).
			Return(&dto.Wallet{
				Name:    testWalletName,
				Balance: testAmount.GetInt(),
			}, nil)
		ts.mockRepo.EXPECT().IncreaseWalletBalance(testWalletName, testAmount.GetInt()).
			Return(nil)

		deposit := dto.Deposit{
			Wallet: testWalletName,
			Amount: testAmount,
		}
		err := ts.svc.IncreaseWalletBalance(deposit)
		require.NoError(t, err)
	})

	t.Run("database connection error", func(t *testing.T) {
		ts := newTestService(t)
		defer ts.Finish()

		ts.mockRepo.EXPECT().GetWallet(testWalletName).
			Return(nil, sql.ErrConnDone)

		deposit := dto.Deposit{
			Wallet: testWalletName,
			Amount: testAmount,
		}
		err := ts.svc.IncreaseWalletBalance(deposit)
		require.Error(t, err)
		e, ok := err.(*httperr.Error)
		require.True(t, ok)
		assert.Equal(t, sql.ErrConnDone, e.Err)
		assert.Equal(t, ErrDatabase.StatusCode, e.StatusCode)
		assert.Equal(t, ErrDatabase.Message, e.Message)
	})

	t.Run("not found", func(t *testing.T) {
		ts := newTestService(t)
		defer ts.Finish()

		ts.mockRepo.EXPECT().GetWallet(testWalletName).
			Return(nil, nil)

		deposit := dto.Deposit{
			Wallet: testWalletName,
			Amount: testAmount,
		}
		err := ts.svc.IncreaseWalletBalance(deposit)
		require.Error(t, err)
		e, ok := err.(*httperr.Error)
		require.True(t, ok)
		assert.Nil(t, e.Err)
		assert.Equal(t, http.StatusNotFound, e.StatusCode)
		assert.Equal(t, "wallet not found", e.Message)
	})
}

// TestService ---------------------------------------------------------------------------------------------------------
type TestService struct {
	t        *testing.T
	log      *zap.SugaredLogger
	mockCtrl *gomock.Controller
	mockRepo *mocks.MockRepository
	svc      *Service
}

func newTestService(t *testing.T) TestService {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	ts := TestService{
		t:        t,
		mockCtrl: mockCtrl,
		mockRepo: mocks.NewMockRepository(mockCtrl),
	}

	if logsEnabled {
		logger, _ := zap.NewDevelopment()
		ts.log = logger.Sugar()
	} else {
		ts.log = zap.NewNop().Sugar()
	}

	ts.svc = NewService(ts.log, ts.mockRepo)

	return ts
}

func (ts *TestService) Finish() {
	ts.mockCtrl.Finish()
}
