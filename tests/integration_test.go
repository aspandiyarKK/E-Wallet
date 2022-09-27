//nolint:bodyclose
package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
	"time"

	"EWallet/pkg/exchange"

	"EWallet/internal"
	"EWallet/internal/rest"
	"EWallet/pkg/repository"

	_ "github.com/jackc/pgx/v4/stdlib"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const pgDSN = "postgres://postgres:secret@localhost:5433/postgres"
const xrHost = "https://api.apilayer.com/exchangerates_data/convert?to="

type IntegrationTestSuite struct {
	suite.Suite
	log      *logrus.Logger
	store    *repository.PG
	router   *rest.Router
	app      *internal.App
	exchange *exchange.Rate
	url      string
}

func (s *IntegrationTestSuite) SetupSuite() {
	ctx := context.Background()
	s.log = logrus.New()
	var err error
	s.store, err = repository.NewRepo(ctx, s.log, pgDSN)
	require.NoError(s.T(), err)
	err = s.store.Migrate(migrate.Up)
	require.NoError(s.T(), err)
	s.exchange = exchange.NewExchangeRate(s.log, xrHost)
	s.app = internal.NewApp(s.log, s.store, s.exchange)
	s.router = rest.NewRouter(s.log, s.app)
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

func (s *IntegrationTestSuite) SetupTest() {
}

func (s *IntegrationTestSuite) TearDownTest() {
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

func (s *IntegrationTestSuite) processRequest(ctx context.Context, method, path string, body interface{}, response interface{}) *http.Response {
	s.T().Helper()
	requestBody, err := json.Marshal(body)
	require.NoError(s.T(), err)
	req, err := http.NewRequestWithContext(ctx, method, path, bytes.NewBuffer(requestBody))
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")
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

func (s *IntegrationTestSuite) TestDepoWallet() {
	ctx := context.Background()
	path := s.url + "/wallet"
	wallet := repository.Wallet{
		Owner:   "test1",
		Balance: 1000,
	}

	finreq := repository.FinRequest{
		Sum: 1000.0,
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

func (s *IntegrationTestSuite) TestWithDrawWallet() {
	ctx := context.Background()
	path := s.url + "/wallet"
	wallet := repository.Wallet{
		Owner:   "test1",
		Balance: 2000,
	}

	finreq := repository.FinRequest{
		Sum: 1000.0,
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

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
