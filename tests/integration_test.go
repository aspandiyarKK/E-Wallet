//nolint:bodyclose
package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	"EWallet/internal"
	"EWallet/internal/rest"
	"EWallet/pkg/repository"

	_ "github.com/jackc/pgx/v4/stdlib"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	pgDSN = "postgres://postgres:secret@localhost:5433/postgres"
	token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFzcGFuIiwiZXhwIjoxNjY4MzM2Mjg5LCJpc3MiOiJlLXdhbGxldCJ9.buETLo3ShqjvXT8Vg5GY2sMOmbLGNRqE3axwht5iN-Y"
)

type IntegrationTestSuite struct {
	suite.Suite
	log    *logrus.Logger
	store  *repository.PG
	router *rest.Router
	app    *internal.App
	url    string
}
type MockExchange struct{}

func (m *MockExchange) GetRate(ctx context.Context, currency string, amount float64) (float64, error) {
	if currency == "usd" {
		return 2, nil
	}
	if currency == "rub" {
		return 1, nil
	}
	if currency == "eur" {
		return 2.25, nil
	}
	if currency == "fjk" {
		return 0, fmt.Errorf("invalid currency")
	}
	return 1, nil
}

func (s *IntegrationTestSuite) SetupSuite() {
	ctx := context.Background()
	s.log = logrus.New()
	var err error
	s.store, err = repository.NewRepo(ctx, s.log, pgDSN)
	require.NoError(s.T(), err)
	err = s.store.Migrate(migrate.Up)
	require.NoError(s.T(), err)
	s.app = internal.NewApp(s.log, s.store, &MockExchange{})
	s.router = rest.NewRouter(s.log, s.app, "testsecret")
	go func() {
		_ = s.router.Run(ctx, "localhost:3001")
	}()
	s.url = "http://localhost:3001/api/v1"
	time.Sleep(100 * time.Millisecond)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	err := s.store.Migrate(migrate.Down)
	require.NoError(s.T(), err)
}

func (s *IntegrationTestSuite) TestGetWalletNotFound() {
	ctx := context.Background()
	wallet := repository.Wallet{
		Owner:   "test1",
		Balance: 1050,
	}
	path := s.url + "/wallet"
	var idMap map[string]int
	resp := s.processRequest(ctx, http.MethodPost, path, wallet, &idMap)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	id, ok := idMap["id"]
	require.True(s.T(), ok)
	var walletResp repository.Wallet
	resp = s.processRequest(ctx, http.MethodGet, path+"/"+strconv.Itoa(id+1), nil, &walletResp)
	require.Equal(s.T(), http.StatusNotFound, resp.StatusCode)
}

func (s *IntegrationTestSuite) TestGetWalletWithRate() {
	ctx := context.Background()
	wallet := repository.Wallet{
		Owner:   "test1",
		Balance: 1050,
	}
	path := s.url + "/wallet"
	var idMap map[string]int
	resp := s.processRequest(ctx, http.MethodPost, path, wallet, &idMap)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	id, ok := idMap["id"]
	require.True(s.T(), ok)

	var walletResp repository.Wallet
	resp = s.processRequest(ctx, http.MethodGet, path+"/"+strconv.Itoa(id)+"?currency=usd", nil, &walletResp)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	require.Equal(s.T(), 2.0, walletResp.Balance)
}

func (s *IntegrationTestSuite) TestCreateAndGetWallet() {
	ctx := context.Background()
	wallet := repository.Wallet{
		Owner:   "test1",
		Balance: 1050,
	}
	path := s.url + "/wallet"
	var idMap map[string]int
	resp := s.processRequest(ctx, http.MethodPost, path, wallet, &idMap)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	id, ok := idMap["id"]
	require.True(s.T(), ok)

	var walletResp repository.Wallet
	resp = s.processRequest(ctx, http.MethodGet, path+"/"+strconv.Itoa(id), nil, &walletResp)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	require.Equal(s.T(), wallet.Owner, walletResp.Owner)
	require.Equal(s.T(), wallet.Balance, walletResp.Balance)
}

func (s *IntegrationTestSuite) TestCreateBadRequest() {
	ctx := context.Background()
	path := s.url + "/wallet"
	resp := s.processRequest(ctx, http.MethodPost, path, " ", nil)
	require.Equal(s.T(), http.StatusBadRequest, resp.StatusCode)
}

func (s *IntegrationTestSuite) TestUpdateWallet() {
	wallet := repository.Wallet{
		Owner:   "test1",
		Balance: 100,
	}
	wallet2 := repository.Wallet{
		Owner:   "test2",
		Balance: 1000,
	}
	ctx := context.Background()
	path := s.url + "/wallet"
	var idMap map[string]int
	resp := s.processRequest(ctx, http.MethodPost, path, wallet, &idMap)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	id, ok := idMap["id"]
	require.True(s.T(), ok)

	var walletResp repository.Wallet
	resp = s.processRequest(ctx, http.MethodPut, path+"/"+strconv.Itoa(id), wallet2, &walletResp)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	require.Equal(s.T(), walletResp.Owner, wallet2.Owner)
	require.Equal(s.T(), walletResp.Balance, wallet2.Balance)
}

func (s *IntegrationTestSuite) TestUpdateWalletNotFound() {
	wallet := repository.Wallet{
		Owner:   "test1",
		Balance: 100,
	}
	wallet2 := repository.Wallet{
		Owner:   "test2",
		Balance: 1000,
	}
	ctx := context.Background()
	path := s.url + "/wallet"
	var idMap map[string]int
	resp := s.processRequest(ctx, http.MethodPost, path, wallet, &idMap)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	id, ok := idMap["id"]
	require.True(s.T(), ok)

	var walletResp repository.Wallet
	resp = s.processRequest(ctx, http.MethodPut, path+"/"+strconv.Itoa(id+1), wallet2, &walletResp)
	require.Equal(s.T(), http.StatusNotFound, resp.StatusCode)
}

func (s *IntegrationTestSuite) TestUpdateWalletBadRequest() {
	wallet := repository.Wallet{
		Owner:   "test1",
		Balance: 100,
	}
	ctx := context.Background()
	path := s.url + "/wallet"
	var idMap map[string]int
	resp := s.processRequest(ctx, http.MethodPost, path, wallet, &idMap)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	id, ok := idMap["id"]
	require.True(s.T(), ok)

	resp = s.processRequest(ctx, http.MethodPut, path+"/"+strconv.Itoa(id), "saksfsklj", nil)
	require.Equal(s.T(), http.StatusBadRequest, resp.StatusCode)
}

func (s *IntegrationTestSuite) TestDeleteWallet() {
	ctx := context.Background()
	wallet := repository.Wallet{
		Owner:   "test1",
		Balance: 100,
	}

	path := s.url + "/wallet"
	var idMap map[string]int
	resp := s.processRequest(ctx, http.MethodPost, path, wallet, &idMap)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	id, ok := idMap["id"]
	require.True(s.T(), ok)

	resp = s.processRequest(ctx, http.MethodDelete, path+"/"+strconv.Itoa(id), nil, nil)

	require.Equal(s.T(), http.StatusNoContent, resp.StatusCode)
}

func (s *IntegrationTestSuite) TestDeleteWalletNotFound() {
	ctx := context.Background()
	wallet := repository.Wallet{
		Owner:   "test1",
		Balance: 100,
	}

	path := s.url + "/wallet"
	var idMap map[string]int
	resp := s.processRequest(ctx, http.MethodPost, path, wallet, &idMap)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	id, ok := idMap["id"]
	require.True(s.T(), ok)

	resp = s.processRequest(ctx, http.MethodDelete, path+"/"+strconv.Itoa(id+1), nil, nil)
	require.Equal(s.T(), http.StatusNotFound, resp.StatusCode)
}

func (s *IntegrationTestSuite) TestDeleteWalletBadRequest() {
	ctx := context.Background()
	wallet := repository.Wallet{
		Owner:   "test1",
		Balance: 100,
	}

	path := s.url + "/wallet"
	var idMap map[string]int
	resp := s.processRequest(ctx, http.MethodPost, path, wallet, &idMap)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	id, ok := idMap["id"]
	require.True(s.T(), ok)

	resp = s.processRequest(ctx, http.MethodDelete, path+"/"+strconv.Itoa(id+1)+"dssdsds", nil, nil)
	require.Equal(s.T(), http.StatusBadRequest, resp.StatusCode)
}

func (s *IntegrationTestSuite) processRequest(ctx context.Context, method, path string, body interface{}, response interface{}) *http.Response {
	s.T().Helper()
	requestBody, err := json.Marshal(body)
	require.NoError(s.T(), err)
	req, err := http.NewRequestWithContext(ctx, method, path, bytes.NewBuffer(requestBody))
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(s.T(), err)
	defer func() {
		require.NoError(s.T(), resp.Body.Close())
	}()
	if response != nil {
		err = json.NewDecoder(resp.Body).Decode(response)
		require.NoError(s.T(), err)
	}
	return resp
}

func (s *IntegrationTestSuite) TestAuth() {
	ctx := context.Background()
	userInfo := rest.UserInfo{
		Username: "aspan",
		Password: "12345",
	}
	path := "http://localhost:3001/auth"

	resp := s.processRequest(ctx, http.MethodPost, path, userInfo, nil)
	require.Equal(s.T(), resp.StatusCode, http.StatusOK)
}

func (s *IntegrationTestSuite) TestAuthBadRequest() {
	ctx := context.Background()
	path := "http://localhost:3001/auth"

	resp := s.processRequest(ctx, http.MethodPost, path, " ", nil)
	require.Equal(s.T(), resp.StatusCode, http.StatusBadRequest)
}

func (s *IntegrationTestSuite) TestAuthUnauthorized() {
	ctx := context.Background()
	path := "http://localhost:3001/auth"
	userInfo := rest.UserInfo{
		Username: "aspan",
		Password: "123457984",
	}
	resp := s.processRequest(ctx, http.MethodPost, path, userInfo, nil)
	require.Equal(s.T(), resp.StatusCode, http.StatusUnauthorized)
}

func (s *IntegrationTestSuite) TestDepoWallet() {
	ctx := context.Background()
	path := s.url + "/wallet"
	wallet := repository.Wallet{
		Owner:   "test1",
		Balance: 1000,
	}

	finreq := repository.FinRequest{
		Sum:  1000.0,
		UUID: "f7eb5a3b-d9d2-11ec-abbd-0242ac150004",
	}
	var idMap map[string]int
	resp := s.processRequest(ctx, http.MethodPost, path, wallet, &idMap)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	id, ok := idMap["id"]
	require.True(s.T(), ok)
	resp = s.processRequest(ctx, http.MethodPut, path+"/"+strconv.Itoa(id)+"/deposit", finreq, nil)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	var walletResp repository.Wallet
	resp = s.processRequest(ctx, http.MethodGet, path+"/"+strconv.Itoa(id), nil, &walletResp)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	require.Equal(s.T(), walletResp.Balance, 2000.0)
}

func (s *IntegrationTestSuite) TestDepoWalletNotFound() {
	ctx := context.Background()
	path := s.url + "/wallet"
	wallet := repository.Wallet{
		Owner:   "test1",
		Balance: 1000,
	}

	finreq := repository.FinRequest{
		Sum:  1000.0,
		UUID: "f7eb5a3b-d9d2-11ec-abbd-0242ac150005",
	}
	var idMap map[string]int
	resp := s.processRequest(ctx, http.MethodPost, path, wallet, &idMap)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	id, ok := idMap["id"]
	require.True(s.T(), ok)
	resp = s.processRequest(ctx, http.MethodPut, path+"/"+strconv.Itoa(id+1)+"/deposit", finreq, nil)
	require.Equal(s.T(), http.StatusNotFound, resp.StatusCode)
}

func (s *IntegrationTestSuite) TestDepoWalletBadRequest() {
	ctx := context.Background()
	path := s.url + "/wallet"
	wallet := repository.Wallet{
		Owner:   "test1",
		Balance: 1000,
	}
	var idMap map[string]int
	resp := s.processRequest(ctx, http.MethodPost, path, wallet, &idMap)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	id, ok := idMap["id"]
	require.True(s.T(), ok)
	resp = s.processRequest(ctx, http.MethodPut, path+"/"+strconv.Itoa(id)+"/deposit", "", nil)
	require.Equal(s.T(), http.StatusBadRequest, resp.StatusCode)
}

func (s *IntegrationTestSuite) TestDepoWalletNonConflict() {
	ctx := context.Background()
	path := s.url + "/wallet"
	wallet := repository.Wallet{
		Owner:   "test1",
		Balance: 1000,
	}

	finreq := repository.FinRequest{
		Sum:  1000.0,
		UUID: "f7eb5a3b-d9d2-11ec-abfd-0242ac170004",
	}
	finreq2 := repository.FinRequest{
		Sum:  1000.0,
		UUID: "f7eb5a3b-d9d2-11ec-abed-0242ac160004",
	}
	var idMap map[string]int
	resp := s.processRequest(ctx, http.MethodPost, path, wallet, &idMap)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	id, ok := idMap["id"]
	require.True(s.T(), ok)
	resp = s.processRequest(ctx, http.MethodPut, path+"/"+strconv.Itoa(id)+"/deposit", finreq, nil)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	var walletResp repository.Wallet
	resp = s.processRequest(ctx, http.MethodGet, path+"/"+strconv.Itoa(id), nil, &walletResp)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	require.Equal(s.T(), walletResp.Balance, 2000.0)

	resp = s.processRequest(ctx, http.MethodPut, path+"/"+strconv.Itoa(id)+"/deposit", finreq2, nil)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	resp = s.processRequest(ctx, http.MethodGet, path+"/"+strconv.Itoa(id), nil, &walletResp)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	require.Equal(s.T(), walletResp.Balance, 3000.0)
}

func (s *IntegrationTestSuite) TestDepoConflict() {
	ctx := context.Background()
	path := s.url + "/wallet"
	wallet := repository.Wallet{
		Owner:   "test1",
		Balance: 1000,
	}

	finreq := repository.FinRequest{
		Sum:  1000.0,
		UUID: "f7eb5a3b-d9d2-11ec-abbd-0242ac110004",
	}
	finreq2 := repository.FinRequest{
		Sum:  1000.0,
		UUID: "f7eb5a3b-d9d2-11ec-abbd-0242ac110004",
	}
	var idMap map[string]int
	resp := s.processRequest(ctx, http.MethodPost, path, wallet, &idMap)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	id, ok := idMap["id"]
	require.True(s.T(), ok)
	resp = s.processRequest(ctx, http.MethodPut, path+"/"+strconv.Itoa(id)+"/deposit", finreq, nil)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	var walletResp repository.Wallet
	resp = s.processRequest(ctx, http.MethodGet, path+"/"+strconv.Itoa(id), nil, &walletResp)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	require.Equal(s.T(), walletResp.Balance, 2000.0)

	resp = s.processRequest(ctx, http.MethodPut, path+"/"+strconv.Itoa(id)+"/deposit", finreq2, nil)
	require.Equal(s.T(), http.StatusConflict, resp.StatusCode)

	resp = s.processRequest(ctx, http.MethodGet, path+"/"+strconv.Itoa(id), nil, &walletResp)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	require.Equal(s.T(), walletResp.Balance, 2000.0)
}

func (s *IntegrationTestSuite) TestWithdrawWallet() {
	ctx := context.Background()
	path := s.url + "/wallet"
	wallet := repository.Wallet{
		Owner:   "test1",
		Balance: 2000,
	}

	finreq := repository.FinRequest{
		Sum:  1000.0,
		UUID: "f7eb5a3b-d9d2-11ec-abbd-0242ac150006",
	}
	var idMap map[string]int
	resp := s.processRequest(ctx, http.MethodPost, path, wallet, &idMap)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	id, ok := idMap["id"]
	require.True(s.T(), ok)
	resp = s.processRequest(ctx, http.MethodPut, path+"/"+strconv.Itoa(id)+"/withdraw", finreq, nil)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	var walletResp repository.Wallet
	resp = s.processRequest(ctx, http.MethodGet, path+"/"+strconv.Itoa(id), nil, &walletResp)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	require.Equal(s.T(), walletResp.Balance, 1000.0)
}

func (s *IntegrationTestSuite) TestWithdrawWalletNotFound() {
	ctx := context.Background()
	path := s.url + "/wallet"
	wallet := repository.Wallet{
		Owner:   "test1",
		Balance: 2000,
	}

	finreq := repository.FinRequest{
		Sum:  1000.0,
		UUID: "f7eb5a3b-d9d2-11ec-abbd-0142ac150007",
	}
	var idMap map[string]int
	resp := s.processRequest(ctx, http.MethodPost, path, wallet, &idMap)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	id, ok := idMap["id"]
	require.True(s.T(), ok)
	resp = s.processRequest(ctx, http.MethodPut, path+"/"+strconv.Itoa(id+1)+"/withdraw", finreq, nil)
	require.Equal(s.T(), http.StatusNotFound, resp.StatusCode)
}

func (s *IntegrationTestSuite) TestWithdrawWalletNonConflict() {
	ctx := context.Background()
	path := s.url + "/wallet"
	wallet := repository.Wallet{
		Owner:   "test1",
		Balance: 2000,
	}

	finreq := repository.FinRequest{
		Sum:  1000.0,
		UUID: "f7eb5a3b-d9d2-11ec-abbd-0132ac150006",
	}
	finreq2 := repository.FinRequest{
		Sum:  1000.0,
		UUID: "f7eb5a3b-d9d2-11ec-abbd-0232ac150006",
	}
	var idMap map[string]int
	resp := s.processRequest(ctx, http.MethodPost, path, wallet, &idMap)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	id, ok := idMap["id"]
	require.True(s.T(), ok)
	resp = s.processRequest(ctx, http.MethodPut, path+"/"+strconv.Itoa(id)+"/withdraw", finreq, nil)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	resp = s.processRequest(ctx, http.MethodPut, path+"/"+strconv.Itoa(id)+"/withdraw", finreq2, nil)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var walletResp repository.Wallet
	resp = s.processRequest(ctx, http.MethodGet, path+"/"+strconv.Itoa(id), nil, &walletResp)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	require.Equal(s.T(), walletResp.Balance, 0.0)
}

func (s *IntegrationTestSuite) TestWithdrawConflict() {
	ctx := context.Background()
	path := s.url + "/wallet"
	wallet := repository.Wallet{
		Owner:   "test1",
		Balance: 2000,
	}

	finreq := repository.FinRequest{
		Sum:  1000.0,
		UUID: "f7eb5a3b-d9d2-11ec-abbd-0192ac150006",
	}
	var idMap map[string]int
	resp := s.processRequest(ctx, http.MethodPost, path, wallet, &idMap)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	id, ok := idMap["id"]
	require.True(s.T(), ok)
	resp = s.processRequest(ctx, http.MethodPut, path+"/"+strconv.Itoa(id)+"/withdraw", finreq, nil)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	resp = s.processRequest(ctx, http.MethodPut, path+"/"+strconv.Itoa(id)+"/withdraw", finreq, nil)
	require.Equal(s.T(), http.StatusConflict, resp.StatusCode)
	var walletResp repository.Wallet
	resp = s.processRequest(ctx, http.MethodGet, path+"/"+strconv.Itoa(id), nil, &walletResp)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	require.Equal(s.T(), walletResp.Balance, 1000.0)
}

func (s *IntegrationTestSuite) TestWithdrawWalletBadRequest() {
	ctx := context.Background()
	path := s.url + "/wallet"
	wallet := repository.Wallet{
		Owner:   "test1",
		Balance: 2000,
	}

	var idMap map[string]int
	resp := s.processRequest(ctx, http.MethodPost, path, wallet, &idMap)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	id, ok := idMap["id"]
	require.True(s.T(), ok)
	resp = s.processRequest(ctx, http.MethodPut, path+"/"+strconv.Itoa(id)+"/withdraw", " ", nil)
	require.Equal(s.T(), http.StatusBadRequest, resp.StatusCode)
}

func (s *IntegrationTestSuite) TestTransferWallet() {
	ctx := context.Background()
	path := s.url + "/wallet"
	wallet := repository.Wallet{
		Owner:   "test1",
		Balance: 1000,
	}
	wallet2 := repository.Wallet{
		Owner:   "test1",
		Balance: 1000,
	}
	var idMap map[string]int
	resp := s.processRequest(ctx, http.MethodPost, path, wallet, &idMap)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	idSender, ok := idMap["id"]
	require.True(s.T(), ok)

	resp = s.processRequest(ctx, http.MethodPost, path, wallet2, &idMap)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	idGetter, ok := idMap["id"]
	require.True(s.T(), ok)
	finreq := repository.FinRequest{
		Sum:          600.0,
		WalletTarget: idGetter,
		UUID:         "f7eb5a3b-d9d2-11ec-abbd-0242ac150008",
	}
	resp = s.processRequest(ctx, http.MethodPut, path+"/"+strconv.Itoa(idSender)+"/transfer", finreq, nil)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var walletRespSender repository.Wallet
	resp = s.processRequest(ctx, http.MethodGet, path+"/"+strconv.Itoa(idSender), nil, &walletRespSender)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	require.Equal(s.T(), walletRespSender.Balance, 400.0)

	var walletRespGetter repository.Wallet
	resp = s.processRequest(ctx, http.MethodGet, path+"/"+strconv.Itoa(idGetter), nil, &walletRespGetter)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	require.Equal(s.T(), walletRespGetter.Balance, 1600.0)
}

func (s *IntegrationTestSuite) TestTransferWalletNonConflict() {
	ctx := context.Background()
	path := s.url + "/wallet"
	wallet := repository.Wallet{
		Owner:   "test1",
		Balance: 1000,
	}
	wallet2 := repository.Wallet{
		Owner:   "test1",
		Balance: 1000,
	}
	var idMap map[string]int
	resp := s.processRequest(ctx, http.MethodPost, path, wallet, &idMap)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	idSender, ok := idMap["id"]
	require.True(s.T(), ok)

	resp = s.processRequest(ctx, http.MethodPost, path, wallet2, &idMap)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	idGetter, ok := idMap["id"]
	require.True(s.T(), ok)
	finreq := repository.FinRequest{
		Sum:          600.0,
		WalletTarget: idGetter,
		UUID:         "f7eb5a3b-d9d2-19ec-abbd-0242ac150008",
	}

	finreq2 := repository.FinRequest{
		Sum:          200.0,
		WalletTarget: idGetter,
		UUID:         "f7eb5a3b-d9d2-02ec-abbd-0242ac150008",
	}
	resp = s.processRequest(ctx, http.MethodPut, path+"/"+strconv.Itoa(idSender)+"/transfer", finreq, nil)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	resp = s.processRequest(ctx, http.MethodPut, path+"/"+strconv.Itoa(idSender)+"/transfer", finreq2, nil)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var walletRespSender repository.Wallet
	resp = s.processRequest(ctx, http.MethodGet, path+"/"+strconv.Itoa(idSender), nil, &walletRespSender)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	require.Equal(s.T(), walletRespSender.Balance, 200.0)

	var walletRespGetter repository.Wallet
	resp = s.processRequest(ctx, http.MethodGet, path+"/"+strconv.Itoa(idGetter), nil, &walletRespGetter)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	require.Equal(s.T(), walletRespGetter.Balance, 1800.0)
}

func (s *IntegrationTestSuite) TestTransferWalletConflict() {
	ctx := context.Background()
	path := s.url + "/wallet"
	wallet := repository.Wallet{
		Owner:   "test1",
		Balance: 1000,
	}
	wallet2 := repository.Wallet{
		Owner:   "test1",
		Balance: 1000,
	}
	var idMap map[string]int
	resp := s.processRequest(ctx, http.MethodPost, path, wallet, &idMap)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	idSender, ok := idMap["id"]
	require.True(s.T(), ok)

	resp = s.processRequest(ctx, http.MethodPost, path, wallet2, &idMap)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	idGetter, ok := idMap["id"]
	require.True(s.T(), ok)
	finreq := repository.FinRequest{
		Sum:          600.0,
		WalletTarget: idGetter,
		UUID:         "f7eb5a3b-d9d2-19ec-abbd-9442ac150008",
	}

	finreq2 := repository.FinRequest{
		Sum:          200.0,
		WalletTarget: idGetter,
		UUID:         "f7eb5a3b-d9d2-19ec-abbd-9442ac150008",
	}
	resp = s.processRequest(ctx, http.MethodPut, path+"/"+strconv.Itoa(idSender)+"/transfer", finreq, nil)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	resp = s.processRequest(ctx, http.MethodPut, path+"/"+strconv.Itoa(idSender)+"/transfer", finreq2, nil)
	require.Equal(s.T(), http.StatusConflict, resp.StatusCode)

	var walletRespSender repository.Wallet
	resp = s.processRequest(ctx, http.MethodGet, path+"/"+strconv.Itoa(idSender), nil, &walletRespSender)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	require.Equal(s.T(), walletRespSender.Balance, 400.0)

	var walletRespGetter repository.Wallet
	resp = s.processRequest(ctx, http.MethodGet, path+"/"+strconv.Itoa(idGetter), nil, &walletRespGetter)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	require.Equal(s.T(), walletRespGetter.Balance, 1600.0)
}

func (s *IntegrationTestSuite) TestTransferWalletNotFound() {
	ctx := context.Background()
	path := s.url + "/wallet"
	wallet := repository.Wallet{
		Owner:   "test1",
		Balance: 1000,
	}
	wallet2 := repository.Wallet{
		Owner:   "test1",
		Balance: 1000,
	}
	var idMap map[string]int
	resp := s.processRequest(ctx, http.MethodPost, path, wallet, &idMap)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	idSender, ok := idMap["id"]
	require.True(s.T(), ok)

	resp = s.processRequest(ctx, http.MethodPost, path, wallet2, &idMap)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	idGetter, ok := idMap["id"]
	require.True(s.T(), ok)
	finreq := repository.FinRequest{
		Sum:          600.0,
		WalletTarget: idGetter,
		UUID:         "f7eb5a3b-d9d2-11ec-abbd-0242ac150009",
	}
	resp = s.processRequest(ctx, http.MethodPut, path+"/"+strconv.Itoa(idSender+2)+"/transfer", finreq, nil)
	require.Equal(s.T(), http.StatusNotFound, resp.StatusCode)
}

func (s *IntegrationTestSuite) TestTransferWalletBadRequest() {
	ctx := context.Background()
	path := s.url + "/wallet"
	wallet := repository.Wallet{
		Owner:   "test1",
		Balance: 1000,
	}
	wallet2 := repository.Wallet{
		Owner:   "test1",
		Balance: 1000,
	}
	var idMap map[string]int
	resp := s.processRequest(ctx, http.MethodPost, path, wallet, &idMap)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	idSender, ok := idMap["id"]
	require.True(s.T(), ok)

	resp = s.processRequest(ctx, http.MethodPost, path, wallet2, &idMap)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	resp = s.processRequest(ctx, http.MethodPut, path+"/"+strconv.Itoa(idSender)+"/transfer", " ", nil)
	require.Equal(s.T(), http.StatusBadRequest, resp.StatusCode)
}

func (s *IntegrationTestSuite) TestGetTransaction() {
	ctx := context.Background()
	path := s.url + "/wallet"
	wallet := repository.Wallet{
		Owner:   "test1",
		Balance: 1000,
	}

	finreq := repository.FinRequest{
		Sum:  1000.0,
		UUID: "f7eb5a3b-d9d2-11ec-abbd-0242ac159934",
	}
	finreq2 := repository.FinRequest{
		Sum:  1000.0,
		UUID: "f7eb5a3b-d9d2-11ec-abbd-0242ac150769",
	}
	var idMap map[string]int
	resp := s.processRequest(ctx, http.MethodPost, path, wallet, &idMap)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	id, ok := idMap["id"]
	require.True(s.T(), ok)

	resp = s.processRequest(ctx, http.MethodPut, path+"/"+strconv.Itoa(id)+"/deposit", finreq, nil)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	resp = s.processRequest(ctx, http.MethodPut, path+"/"+strconv.Itoa(id)+"/deposit", finreq2, nil)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var ans []repository.Transaction
	path = s.url + "/transaction"
	resp = s.processRequest(ctx, http.MethodGet, path+"/"+strconv.Itoa(id), nil, &ans)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	require.Equal(s.T(), len(ans), 3)
}

func (s *IntegrationTestSuite) TestGetTransactionBySum() {
	ctx := context.Background()
	path := s.url + "/wallet"
	wallet := repository.Wallet{
		Owner:   "test1",
		Balance: 1000,
	}

	finreq := repository.FinRequest{
		Sum:  1000.0,
		UUID: "f7eb5a3b-d9d2-11ec-abbd-0242ac019934",
	}
	finreq2 := repository.FinRequest{
		Sum:  3000.0,
		UUID: "f7eb5a3b-d9d2-11ec-abbd-0242ac819934",
	}

	var idMap map[string]int
	resp := s.processRequest(ctx, http.MethodPost, path, wallet, &idMap)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	id, ok := idMap["id"]
	require.True(s.T(), ok)

	resp = s.processRequest(ctx, http.MethodPut, path+"/"+strconv.Itoa(id)+"/deposit", finreq, nil)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	resp = s.processRequest(ctx, http.MethodPut, path+"/"+strconv.Itoa(id)+"/deposit", finreq2, nil)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var ans []repository.Transaction
	// TODO: url
	path = s.url + "/transaction"
	resp = s.processRequest(ctx, http.MethodGet, path+"/"+strconv.Itoa(id)+"?order=sum", nil, &ans)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	require.Equal(s.T(), ans[0].Sum, finreq2.Sum)
	require.Equal(s.T(), ans[1].Sum, finreq.Sum)
}

// order by ... asc/desc, limit + offset

func (s *IntegrationTestSuite) TestGetTransactionNotFound() {
	ctx := context.Background()
	path := s.url + "/wallet"
	wallet := repository.Wallet{
		Owner:   "test1",
		Balance: 1000,
	}

	finreq := repository.FinRequest{
		Sum:  1000.0,
		UUID: "f7eb5a3b-d9d2-11ec-abbd-0242ac219934",
	}

	var idMap map[string]int
	resp := s.processRequest(ctx, http.MethodPost, path, wallet, &idMap)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	id, ok := idMap["id"]
	require.True(s.T(), ok)

	resp = s.processRequest(ctx, http.MethodPut, path+"/"+strconv.Itoa(id)+"/deposit", finreq, nil)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	path = s.url + "/transaction"
	resp = s.processRequest(ctx, http.MethodGet, path+"/"+strconv.Itoa(id+1), nil, nil)
	require.Equal(s.T(), http.StatusNotFound, resp.StatusCode)
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
