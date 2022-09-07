package tests

import (
	"EWallet/internal/rest"
	"EWallet/pkg/repository"
	"bytes"
	"context"
	"encoding/json"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"net/http"
	"strconv"
	"testing"
	"time"
)

const pgDSN = "postgres://postgres:secret@localhost:5433/postgres"

type IntegrationTestSuite struct {
	suite.Suite
	log    *logrus.Logger
	store  *repository.PG
	router *rest.Router
	url    string
}

func (s *IntegrationTestSuite) SetupSuite() {
	ctx := context.Background()
	s.log = logrus.New()
	var err error
	s.store, err = repository.NewRepo(ctx, s.log, pgDSN)
	require.NoError(s.T(), err)
	s.router = rest.NewRouter(s.log, s.store)
	go func() {
		_ = s.router.Run(ctx, "localhost:3001")
	}()
	s.url = "http://localhost:3001/api/v1"
	time.Sleep(100 * time.Millisecond)
}

func (s *IntegrationTestSuite) TearDownSuite() {
}

func (s *IntegrationTestSuite) SetupTest() {
}

func (s *IntegrationTestSuite) TearDownTest() {
}

func (s *IntegrationTestSuite) TestCreateAndGetWallet() {
	wallet := repository.Wallet{
		Owner:   "test1",
		Balance: 1050,
	}
	requestBody, err := json.Marshal(wallet)
	require.NoError(s.T(), err)
	path := s.url + "/wallet"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, path, bytes.NewBuffer(requestBody))
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	var idMap map[string]int
	id, ok := idMap["id"]

	require.True(s.T(), ok)
	err = json.NewDecoder(resp.Body).Decode(&idMap)
	require.NoError(s.T(), err)
	req, err = http.NewRequestWithContext(context.Background(), http.MethodGet, path+"/"+strconv.Itoa(id), nil)
	require.NoError(s.T(), err)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	var walletResp repository.Wallet
	err = json.NewDecoder(resp.Body).Decode(&walletResp)
	require.NoError(s.T(), err)
	require.Equal(s.T(), wallet.Owner, walletResp.Owner)
	require.Equal(s.T(), wallet.Balance, walletResp.Balance)
}

func (s *IntegrationTestSuite) processRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	panic("implement me")
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
