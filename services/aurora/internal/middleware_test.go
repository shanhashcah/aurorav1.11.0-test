//lint:file-ignore U1001 Ignore all unused code, staticcheck doesn't understand testify/suite
package aurora

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/hcnet/throttled"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/hcnet/go/services/aurora/internal/actions"
	auroraContext "github.com/hcnet/go/services/aurora/internal/context"
	"github.com/hcnet/go/services/aurora/internal/db2/history"
	"github.com/hcnet/go/services/aurora/internal/httpx"
	"github.com/hcnet/go/services/aurora/internal/ingest"
	"github.com/hcnet/go/services/aurora/internal/ledger"
	hProblem "github.com/hcnet/go/services/aurora/internal/render/problem"
	"github.com/hcnet/go/services/aurora/internal/test"
	"github.com/hcnet/go/support/db"
	"github.com/hcnet/go/support/log"
	"github.com/hcnet/go/xdr"
)

func requestHelperRemoteAddr(ip string) func(r *http.Request) {
	return func(r *http.Request) {
		r.RemoteAddr = ip
	}
}

func requestHelperXFF(xff string) func(r *http.Request) {
	return func(r *http.Request) {
		r.Header.Set("X-Forwarded-For", xff)
	}
}

type RateLimitMiddlewareTestSuite struct {
	suite.Suite
	ht  *HTTPT
	c   Config
	app *App
	rh  test.RequestHelper
}

func (suite *RateLimitMiddlewareTestSuite) SetupSuite() {
	suite.ht = StartHTTPTest(suite.T(), "base")
}

func (suite *RateLimitMiddlewareTestSuite) SetupTest() {
	suite.c = NewTestConfig()
	suite.c.RateQuota = &throttled.RateQuota{
		MaxRate:  throttled.PerHour(10),
		MaxBurst: 9,
	}
	app, err := NewApp(suite.c)
	if err != nil {
		log.Fatal("cannot initialize app", err)
	}
	suite.app = app
	suite.rh = NewRequestHelper(suite.app)
}

func (suite *RateLimitMiddlewareTestSuite) TearDownSuite() {
	suite.ht.Finish()
}

func (suite *RateLimitMiddlewareTestSuite) TearDownTest() {
	suite.app.Close()
}

// Sets X-RateLimit-Limit headers correctly.
func (suite *RateLimitMiddlewareTestSuite) TestRateLimit_LimitHeaders() {
	w := suite.rh.Get("/")
	assert.Equal(suite.T(), 200, w.Code)
	assert.Equal(suite.T(), "10", w.Header().Get("X-RateLimit-Limit"))
}

// Sets X-RateLimit-Remaining headers correctly.
func (suite *RateLimitMiddlewareTestSuite) TestRateLimit_RemainingHeaders() {
	for i := 0; i < 10; i++ {
		w := suite.rh.Get("/")
		expected := 10 - (i + 1)
		assert.Equal(suite.T(), strconv.Itoa(expected), w.Header().Get("X-RateLimit-Remaining"))
	}

	// confirm remaining stays at 0
	for i := 0; i < 10; i++ {
		w := suite.rh.Get("/")
		assert.Equal(suite.T(), "0", w.Header().Get("X-RateLimit-Remaining"))
	}
}

// Sets X-RateLimit-Reset header correctly. Should reset after 360 seconds since it's limited to 10 requests/hour.
func (suite *RateLimitMiddlewareTestSuite) TestRateLimit_ResetHeaders() {
	w := suite.rh.Get("/")
	assert.Equal(suite.T(), "360", w.Header().Get("X-RateLimit-Reset"))
}

// Restricts based on RemoteAddr IP after too many requests.
func (suite *RateLimitMiddlewareTestSuite) TestRateLimit_RemoteAddr() {
	for i := 0; i < 10; i++ {
		w := suite.rh.Get("/")
		assert.Equal(suite.T(), 200, w.Code)
	}

	w := suite.rh.Get("/")
	assert.Equal(suite.T(), 429, w.Code)

	w = suite.rh.Get("/", requestHelperRemoteAddr("127.0.0.2"))
	assert.Equal(suite.T(), 200, w.Code)

	// Ignores ports
	w = suite.rh.Get("/", requestHelperRemoteAddr("127.0.0.1:4312"))
	assert.Equal(suite.T(), 429, w.Code)
}

// Restrict based upon X-Forwarded-For correctly.
func (suite *RateLimitMiddlewareTestSuite) TestRateLimit_XForwardedFor() {
	for i := 0; i < 10; i++ {
		w := suite.rh.Get("/", requestHelperXFF("4.4.4.4"))
		assert.Equal(suite.T(), 200, w.Code)
	}

	w := suite.rh.Get("/", requestHelperXFF("4.4.4.4"))
	assert.Equal(suite.T(), 429, w.Code)

	// allow other ips
	w = suite.rh.Get("/", requestHelperRemoteAddr("4.4.4.3"))
	assert.Equal(suite.T(), 200, w.Code)

	// Ignores leading private ips
	w = suite.rh.Get("/", requestHelperXFF("10.0.0.1, 4.4.4.4"))
	assert.Equal(suite.T(), 429, w.Code)

	// Ignores trailing ips
	w = suite.rh.Get("/", requestHelperXFF("4.4.4.4, 4.4.4.5, 127.0.0.1"))
	assert.Equal(suite.T(), 429, w.Code)
}

func TestRateLimitMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, new(RateLimitMiddlewareTestSuite))
}

func TestStateMiddleware(t *testing.T) {
	tt := test.Start(t)
	defer tt.Finish()
	test.ResetAuroraDB(t, tt.AuroraDB)

	q := &history.Q{tt.AuroraSession()}

	request, err := http.NewRequest("GET", "http://localhost", nil)
	tt.Assert.NoError(err)

	expectTransaction := true
	endpoint := func(w http.ResponseWriter, r *http.Request) {
		session := r.Context().Value(&auroraContext.SessionContextKey).(*db.Session)
		if (session.GetTx() == nil) == expectTransaction {
			t.Fatalf("expected transaction to be in session: %v", expectTransaction)
		}
		w.WriteHeader(http.StatusOK)
	}

	stateMiddleware := &httpx.StateMiddleware{
		AuroraSession: tt.AuroraSession(),
	}
	handler := stateMiddleware.Wrap(http.HandlerFunc(endpoint))

	for i, testCase := range []struct {
		name                string
		noStateVerification bool
		stateInvalid        bool
		latestHistoryLedger xdr.Uint32
		lastIngestedLedger  uint32
		ingestionVersion    int
		sseRequest          bool
		expectedStatus      int
		expectTransaction   bool
	}{
		{
			name:                "responds with 500 if q.GetExpStateInvalid returns true",
			stateInvalid:        true,
			latestHistoryLedger: 2,
			lastIngestedLedger:  2,
			ingestionVersion:    ingest.CurrentVersion,
			sseRequest:          false,
			expectedStatus:      http.StatusInternalServerError,
			expectTransaction:   false,
		},
		{
			name:                "responds with still ingesting if lastIngestedLedger <= 0",
			stateInvalid:        false,
			latestHistoryLedger: 1,
			lastIngestedLedger:  0,
			ingestionVersion:    ingest.CurrentVersion,
			sseRequest:          false,
			expectedStatus:      hProblem.StillIngesting.Status,
			expectTransaction:   false,
		},
		{
			name:                "responds with still ingesting if lastIngestedLedger < latestHistoryLedger",
			stateInvalid:        false,
			latestHistoryLedger: 3,
			lastIngestedLedger:  2,
			ingestionVersion:    ingest.CurrentVersion,
			sseRequest:          false,
			expectedStatus:      hProblem.StillIngesting.Status,
			expectTransaction:   false,
		},
		{
			name:                "responds with still ingesting if lastIngestedLedger > latestHistoryLedger",
			stateInvalid:        false,
			latestHistoryLedger: 4,
			lastIngestedLedger:  5,
			ingestionVersion:    ingest.CurrentVersion,
			sseRequest:          false,
			expectedStatus:      hProblem.StillIngesting.Status,
			expectTransaction:   false,
		},
		{
			name:                "responds with still ingesting if version != ingest.CurrentVersion",
			stateInvalid:        false,
			latestHistoryLedger: 5,
			lastIngestedLedger:  5,
			ingestionVersion:    ingest.CurrentVersion - 1,
			sseRequest:          false,
			expectedStatus:      hProblem.StillIngesting.Status,
			expectTransaction:   false,
		},
		{
			name:                "succeeds",
			stateInvalid:        false,
			latestHistoryLedger: 6,
			lastIngestedLedger:  6,
			ingestionVersion:    ingest.CurrentVersion,
			sseRequest:          false,
			expectedStatus:      http.StatusOK,
			expectTransaction:   true,
		},
		{
			name:                "succeeds with SSE request",
			stateInvalid:        false,
			latestHistoryLedger: 7,
			lastIngestedLedger:  7,
			ingestionVersion:    ingest.CurrentVersion,
			sseRequest:          true,
			expectedStatus:      http.StatusOK,
			expectTransaction:   false,
		},
		{
			name:                "succeeds without state verification",
			noStateVerification: true,
			stateInvalid:        false,
			latestHistoryLedger: 8,
			lastIngestedLedger:  8,
			ingestionVersion:    ingest.CurrentVersion,
			sseRequest:          false,
			expectedStatus:      http.StatusOK,
			expectTransaction:   true,
		},
		{
			name:                "succeeds without state verification and invalid state",
			noStateVerification: true,
			stateInvalid:        true,
			latestHistoryLedger: 9,
			lastIngestedLedger:  9,
			ingestionVersion:    ingest.CurrentVersion,
			sseRequest:          false,
			expectedStatus:      http.StatusOK,
			expectTransaction:   true,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			stateMiddleware.NoStateVerification = testCase.noStateVerification
			tt.Assert.NoError(q.UpdateExpStateInvalid(testCase.stateInvalid))
			_, err = q.InsertLedger(xdr.LedgerHeaderHistoryEntry{
				Hash: xdr.Hash{byte(i)},
				Header: xdr.LedgerHeader{
					LedgerSeq:          testCase.latestHistoryLedger,
					PreviousLedgerHash: xdr.Hash{byte(i)},
				},
			}, 0, 0, 0, 0, 0)
			tt.Assert.NoError(err)
			tt.Assert.NoError(q.UpdateLastLedgerExpIngest(testCase.lastIngestedLedger))
			tt.Assert.NoError(q.UpdateExpIngestVersion(testCase.ingestionVersion))

			if testCase.sseRequest {
				request.Header.Set("Accept", "text/event-stream")
			} else {
				request.Header.Del("Accept")
			}

			w := httptest.NewRecorder()
			expectTransaction = testCase.expectTransaction
			handler.ServeHTTP(w, request)
			tt.Assert.Equal(testCase.expectedStatus, w.Code)
			if testCase.expectedStatus == http.StatusOK && !testCase.sseRequest {
				tt.Assert.Equal(
					w.Header().Get(actions.LastLedgerHeaderName),
					strconv.FormatInt(int64(testCase.lastIngestedLedger), 10))
			} else {
				tt.Assert.Equal(w.Header().Get(actions.LastLedgerHeaderName), "")
			}
		})
	}
}

func TestCheckHistoryStaleMiddleware(t *testing.T) {
	tt := test.Start(t)
	defer tt.Finish()
	request, err := http.NewRequest("GET", "http://localhost", nil)
	tt.Assert.NoError(err)

	endpoint := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	for _, testCase := range []struct {
		name           string
		coreLatest     int32
		historyLatest  int32
		expectedStatus int
		staleThreshold int32
	}{
		{
			name:           "responds with a service unavailable if history is stale",
			coreLatest:     4,
			historyLatest:  2,
			expectedStatus: http.StatusServiceUnavailable,
			staleThreshold: 1,
		},
		{
			name:           "succeeds",
			coreLatest:     6,
			historyLatest:  6,
			expectedStatus: http.StatusOK,
			staleThreshold: 1,
		},
		{
			name:           "succeeds with threshold 0",
			coreLatest:     6,
			historyLatest:  5,
			expectedStatus: http.StatusOK,
			staleThreshold: 0,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			state := ledger.State{
				CoreLatest:    testCase.coreLatest,
				HistoryLatest: testCase.historyLatest,
			}
			ledger.SetState(state)
			historyMiddleware := httpx.NewHistoryMiddleware(testCase.staleThreshold, tt.AuroraSession())
			handler := historyMiddleware(http.HandlerFunc(endpoint))
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, request)
			tt.Assert.Equal(testCase.expectedStatus, w.Code)
		})
	}
}
