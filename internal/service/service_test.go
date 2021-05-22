package service

import (
	"database/sql"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/ezhdanovskiy/wallets/internal/consts"
	"github.com/ezhdanovskiy/wallets/internal/dto"
	"github.com/ezhdanovskiy/wallets/internal/service/mocks"
)

const logsEnabled = false

const (
	testWalletName01            = "WalletName01"
	testWalletName02            = "WalletName02"
	testAmount       dto.Amount = 123.45
)

func TestService_CreateWallet(t *testing.T) {
	t.Run("empty wallet name", func(t *testing.T) {
		ts := newTestService(t)
		defer ts.Finish()

		req := dto.CreateWalletRequest{Name: ""}
		err := ts.svc.CreateWallet(req)
		assert.Equal(t, ErrEmptyWalletName, err)
	})

	t.Run("database connection error", func(t *testing.T) {
		ts := newTestService(t)
		defer ts.Finish()

		ts.mockRepo.EXPECT().CreateWallet(testWalletName01).
			Return(sql.ErrConnDone)

		req := dto.CreateWalletRequest{Name: testWalletName01}
		err := ts.svc.CreateWallet(req)
		assert.Equal(t, ErrDatabase.Wrap(sql.ErrConnDone), err)
	})

	t.Run("success", func(t *testing.T) {
		ts := newTestService(t)
		defer ts.Finish()

		ts.mockRepo.EXPECT().CreateWallet(testWalletName01).
			Return(nil)

		req := dto.CreateWalletRequest{Name: testWalletName01}
		err := ts.svc.CreateWallet(req)
		assert.NoError(t, err)
	})
}

func TestService_IncreaseWalletBalance(t *testing.T) {
	t.Run("empty wallet name", func(t *testing.T) {
		ts := newTestService(t)
		defer ts.Finish()

		deposit := dto.Deposit{
			Wallet: "",
			Amount: testAmount,
		}
		err := ts.svc.IncreaseWalletBalance(deposit)
		assert.Equal(t, ErrEmptyWalletName, err)
	})

	t.Run("negative amount", func(t *testing.T) {
		ts := newTestService(t)
		defer ts.Finish()

		deposit := dto.Deposit{
			Wallet: testWalletName01,
			Amount: -1,
		}
		err := ts.svc.IncreaseWalletBalance(deposit)
		assert.Equal(t, ErrNotPositiveAmount, err)
	})

	t.Run("database connection error", func(t *testing.T) {
		ts := newTestService(t)
		defer ts.Finish()

		ts.mockRepo.EXPECT().GetWallet(testWalletName01).
			Return(nil, sql.ErrConnDone)

		deposit := dto.Deposit{
			Wallet: testWalletName01,
			Amount: testAmount,
		}
		err := ts.svc.IncreaseWalletBalance(deposit)
		assert.Equal(t, ErrDatabase.Wrap(sql.ErrConnDone), err)
	})

	t.Run("not found", func(t *testing.T) {
		ts := newTestService(t)
		defer ts.Finish()

		ts.mockRepo.EXPECT().GetWallet(testWalletName01).
			Return(nil, nil)

		deposit := dto.Deposit{
			Wallet: testWalletName01,
			Amount: testAmount,
		}
		err := ts.svc.IncreaseWalletBalance(deposit)
		assert.Equal(t, ErrWalletNotFound, err)
	})

	t.Run("success", func(t *testing.T) {
		ts := newTestService(t)
		defer ts.Finish()

		ts.mockRepo.EXPECT().GetWallet(testWalletName01).
			Return(&dto.Wallet{
				Name:    testWalletName01,
				Balance: testAmount.GetInt(),
			}, nil)
		ts.mockRepo.EXPECT().IncreaseWalletBalance(testWalletName01, testAmount.GetInt()).
			Return(nil)

		deposit := dto.Deposit{
			Wallet: testWalletName01,
			Amount: testAmount,
		}
		err := ts.svc.IncreaseWalletBalance(deposit)
		require.NoError(t, err)
	})
}

func TestService_Transfer(t *testing.T) {
	t.Run("empty wallet_from", func(t *testing.T) {
		ts := newTestService(t)
		defer ts.Finish()

		transfer := dto.Transfer{
			WalletFrom: "",
			WalletTo:   testWalletName02,
			Amount:     testAmount,
		}
		err := ts.svc.Transfer(transfer)
		assert.Equal(t, ErrEmptyWalletFrom, err)
	})

	t.Run("empty wallet_to", func(t *testing.T) {
		ts := newTestService(t)
		defer ts.Finish()

		transfer := dto.Transfer{
			WalletFrom: testWalletName01,
			WalletTo:   "",
			Amount:     testAmount,
		}
		err := ts.svc.Transfer(transfer)
		assert.Equal(t, ErrEmptyWalletTo, err)
	})

	t.Run("same wallets", func(t *testing.T) {
		ts := newTestService(t)
		defer ts.Finish()

		transfer := dto.Transfer{
			WalletFrom: testWalletName01,
			WalletTo:   testWalletName01,
			Amount:     testAmount,
		}
		err := ts.svc.Transfer(transfer)
		assert.Equal(t, ErrSameWallets, err)
	})

	t.Run("negative amount", func(t *testing.T) {
		ts := newTestService(t)
		defer ts.Finish()

		transfer := dto.Transfer{
			WalletFrom: testWalletName01,
			WalletTo:   testWalletName02,
			Amount:     -1,
		}
		err := ts.svc.Transfer(transfer)
		assert.Equal(t, ErrNotPositiveAmount, err)
	})
}

func TestService_GetOperations(t *testing.T) {
	t.Run("empty wallet name", func(t *testing.T) {
		ts := newTestService(t)
		defer ts.Finish()

		ops, err := ts.svc.GetOperations(dto.OperationsFilter{})
		assert.Nil(t, ops)
		assert.Equal(t, ErrEmptyWalletName, err)
	})

	t.Run("unsupported operation type", func(t *testing.T) {
		ts := newTestService(t)
		defer ts.Finish()

		filter := dto.OperationsFilter{
			Wallet: testWalletName01,
			Type:   "123",
		}
		ops, err := ts.svc.GetOperations(filter)
		assert.Nil(t, ops)
		assert.Equal(t, ErrUnsupportedOperationType, err)
	})

	t.Run("negative start_date", func(t *testing.T) {
		ts := newTestService(t)
		defer ts.Finish()

		filter := dto.OperationsFilter{
			Wallet:    testWalletName01,
			StartDate: -1,
		}
		ops, err := ts.svc.GetOperations(filter)
		assert.Nil(t, ops)
		assert.Equal(t, ErrNegativeStartDate, err)
	})

	t.Run("negative end_date", func(t *testing.T) {
		ts := newTestService(t)
		defer ts.Finish()

		filter := dto.OperationsFilter{
			Wallet:  testWalletName01,
			EndDate: -1,
		}
		ops, err := ts.svc.GetOperations(filter)
		assert.Nil(t, ops)
		assert.Equal(t, ErrNegativeEndDate, err)
	})

	t.Run("negative limit", func(t *testing.T) {
		ts := newTestService(t)
		defer ts.Finish()

		filter := dto.OperationsFilter{
			Wallet: testWalletName01,
			Limit:  -1,
		}
		ops, err := ts.svc.GetOperations(filter)
		assert.Nil(t, ops)
		assert.Equal(t, ErrNotPositiveLimit, err)
	})

	t.Run("negative offset", func(t *testing.T) {
		ts := newTestService(t)
		defer ts.Finish()

		filter := dto.OperationsFilter{
			Wallet: testWalletName01,
			Offset: -1,
		}
		ops, err := ts.svc.GetOperations(filter)
		assert.Nil(t, ops)
		assert.Equal(t, ErrNegativeOffset, err)
	})

	t.Run("database connection error", func(t *testing.T) {
		ts := newTestService(t)
		defer ts.Finish()

		ts.mockRepo.EXPECT().GetOperations(dto.OperationsFilter{
			Wallet: testWalletName01,
			Limit:  consts.DefaultOperationsLimit,
		}).
			Return(nil, sql.ErrConnDone)

		ops, err := ts.svc.GetOperations(dto.OperationsFilter{
			Wallet: testWalletName01,
		})
		assert.Nil(t, ops)
		assert.Equal(t, ErrDatabase.Wrap(sql.ErrConnDone), err)
	})

	t.Run("success", func(t *testing.T) {
		ts := newTestService(t)
		defer ts.Finish()

		ts.mockRepo.EXPECT().GetOperations(dto.OperationsFilter{
			Wallet: testWalletName01,
			Limit:  consts.DefaultOperationsLimit,
		}).
			Return([]dto.Operation{{
				Wallet: testWalletName01,
				Amount: testAmount,
				Type:   consts.OperationTypeDeposit,
			}}, nil)

		ops, err := ts.svc.GetOperations(dto.OperationsFilter{
			Wallet: testWalletName01,
		})
		assert.NoError(t, err)
		require.Len(t, ops, 1)
		assert.Equal(t, testWalletName01, ops[0].Wallet)
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
