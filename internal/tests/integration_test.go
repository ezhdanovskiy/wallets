package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/ezhdanovskiy/wallets/internal/dto"
	httpsrv "github.com/ezhdanovskiy/wallets/internal/http"
	"github.com/ezhdanovskiy/wallets/internal/repository"
	"github.com/ezhdanovskiy/wallets/internal/service"
)

const logsEnabled = false

func TestCreateWallet(t *testing.T) {
	ts := newTestService(t)
	defer ts.Finish()

	const testWalletName = "TestCreateWalletName01"

	code, body := ts.doRequest(http.MethodPost, "/wallets", dto.CreateWalletRequest{Name: testWalletName})
	assert.Equal(t, 200, code)
	assert.Empty(t, body)

	wallet, err := ts.repo.GetWallet(testWalletName)
	require.NoError(t, err)
	require.NotNil(t, wallet)
	assert.EqualValues(t, 0, wallet.Balance)

	ts.cleanWallets(testWalletName)
}

func TestDeposit(t *testing.T) {
	ts := newTestService(t)
	defer ts.Finish()

	const (
		testWalletName            = "TestDepositWalletName01"
		testAmount     dto.Amount = 123.45
	)
	ts.cleanWallets(testWalletName)

	code, body := ts.doRequest(http.MethodPost, "/wallets", dto.CreateWalletRequest{Name: testWalletName})
	assert.Equal(t, 200, code)

	code, body = ts.doRequest(http.MethodPost, "/wallets/deposit", dto.Deposit{
		Wallet: testWalletName,
		Amount: testAmount,
	})
	assert.Equal(t, 200, code)
	assert.Empty(t, body)

	wallet, err := ts.repo.GetWallet(testWalletName)
	require.NoError(t, err)
	require.NotNil(t, wallet)
	assert.EqualValues(t, testAmount.GetInt(), wallet.Balance)

	operations, err := ts.repo.GetOperations(dto.OperationsFilter{Wallet: testWalletName})
	require.NoError(t, err)
	require.Len(t, operations, 1)
	assert.Equal(t, testWalletName, operations[0].Wallet)
	assert.Equal(t, repository.OperationTypeDeposit, operations[0].Type)
	assert.Equal(t, testAmount, operations[0].Amount)
	assert.Equal(t, repository.SystemWalletName, operations[0].OtherWallet)

	ts.cleanWallets(testWalletName)
}

func TestTransfer(t *testing.T) {
	ts := newTestService(t)
	defer ts.Finish()

	const (
		testWalletName01            = "TestTransferWalletName01"
		testWalletName02            = "TestTransferWalletName02"
		testAmount       dto.Amount = 123.45
	)
	ts.cleanWallets(testWalletName01, testWalletName02)

	require.NoError(t, ts.repo.CreateWallet(testWalletName01))
	require.NoError(t, ts.repo.IncreaseWalletBalance(testWalletName01, testAmount.GetInt()))
	require.NoError(t, ts.repo.CreateWallet(testWalletName02))

	code, body := ts.doRequest(http.MethodPost, "/wallets/transfer", dto.Transfer{
		WalletFrom: testWalletName01,
		WalletTo:   testWalletName02,
		Amount:     testAmount,
	})
	assert.Equal(t, 200, code)
	assert.Empty(t, body)

	wallet, err := ts.repo.GetWallet(testWalletName01)
	require.NoError(t, err)
	require.NotNil(t, wallet)
	assert.EqualValues(t, 0, wallet.Balance)

	wallet, err = ts.repo.GetWallet(testWalletName02)
	require.NoError(t, err)
	require.NotNil(t, wallet)
	assert.EqualValues(t, testAmount.GetInt(), wallet.Balance)

	operations, err := ts.repo.GetOperations(dto.OperationsFilter{Wallet: testWalletName01})
	require.NoError(t, err)
	require.Len(t, operations, 2)
	assert.Equal(t, testWalletName01, operations[0].Wallet)
	assert.Equal(t, repository.OperationTypeDeposit, operations[0].Type)
	assert.Equal(t, testAmount, operations[0].Amount)
	assert.Equal(t, repository.SystemWalletName, operations[0].OtherWallet)
	assert.Equal(t, testWalletName01, operations[1].Wallet)
	assert.Equal(t, repository.OperationTypeWithdrawal, operations[1].Type)
	assert.Equal(t, testAmount, operations[1].Amount)
	assert.Equal(t, testWalletName02, operations[1].OtherWallet)

	operations, err = ts.repo.GetOperations(dto.OperationsFilter{Wallet: testWalletName02})
	require.NoError(t, err)
	require.Len(t, operations, 1)
	assert.Equal(t, testWalletName02, operations[0].Wallet)
	assert.Equal(t, repository.OperationTypeDeposit, operations[0].Type)
	assert.Equal(t, testAmount, operations[0].Amount)
	assert.Equal(t, testWalletName01, operations[0].OtherWallet)

	ts.cleanWallets(testWalletName01, testWalletName02)
}

// TestServer ---------------------------------------------------------------------------------------------------------
type TestServer struct {
	t      *testing.T
	log    *zap.SugaredLogger
	db     *sqlx.DB
	repo   *repository.Repo
	svc    *service.Service
	router *chi.Mux
}

func newTestService(t *testing.T) TestServer {
	t.Parallel()

	var log *zap.SugaredLogger
	if logsEnabled {
		logger, _ := zap.NewDevelopment()
		log = logger.Sugar()
	} else {
		log = zap.NewNop().Sugar()
	}

	dsn := "host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable"
	db, err := sqlx.Connect("postgres", dsn)
	require.NoError(t, err)

	require.NoError(t, repository.MigrateUp(log, db, "file://../../migrations"))

	repo, err := repository.NewRepoWithDB(log, db)
	require.NoError(t, err)

	svc := service.NewService(log, repo)
	srv := httpsrv.NewServer(log, 0, svc)
	router := chi.NewMux()
	router.Group(srv.GetV1ApiRouters())

	ts := TestServer{
		t:      t,
		log:    log,
		db:     db,
		repo:   repo,
		svc:    service.NewService(log, repo),
		router: router,
	}

	return ts
}

func (ts *TestServer) doRequest(method, target string, body interface{}) (code int, respBody string) {
	b := new(bytes.Buffer)
	if str, ok := body.(string); ok {
		b.WriteString(str)
	} else {
		err := json.NewEncoder(b).Encode(body)
		require.NoError(ts.t, err)
	}

	req := httptest.NewRequest(method, target, b)

	recorder := httptest.NewRecorder()
	ts.router.ServeHTTP(recorder, req)

	return recorder.Code, recorder.Body.String()
}

func (ts *TestServer) cleanWallets(wallets ...string) {
	query, args, err := sqlx.In("DELETE FROM wallets WHERE name IN (?)", wallets)
	require.NoError(ts.t, err)

	_, err = ts.db.Exec(ts.db.Rebind(query), args...)
	require.NoError(ts.t, err)

	query, args, err = sqlx.In("DELETE FROM operations WHERE wallet IN (?)", wallets)
	require.NoError(ts.t, err)

	_, err = ts.db.Exec(ts.db.Rebind(query), args...)
	require.NoError(ts.t, err)
}

func (ts *TestServer) Finish() {
}
