// +build unit

package client

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"
)

var (
	accountOne = Account{
		ID:             "ad27e265-9605-4b4b-a0e5-3003ea9cc4dc",
		OrganisationID: "eb0bd6f5-c3f5-44b2-b677-acd23cdde73c",
		Type:           "accounts",
		Version:        0,
		Attributes: Attributes{
			Country:         "GB",
			BaseCurrency:    "GBP",
			AccountNumber:   "41426819",
			BankID:          "400300",
			BankIDCode:      "GBDSC",
			BIC:             "NWBKGB22",
			IBAN:            "GB11NWBK40030041426819",
			Title:           "Ms",
			FirstName:       "Samantha",
			BankAccountName: "Samantha Holder",
			AlternativeBankAccountNames: []string{
				"Sam Holder",
			},
			AccountClassification:   "Personal",
			JointAccount:            false,
			AccountMatchingOptOut:   false,
			SecondaryIdentification: "A1B2C3D4",
		},
	}

	accountTwo = Account{
		ID:             "ad27e265-9605-4b4b-a0e5-3003ea9cc4dd",
		OrganisationID: "eb0bd6f5-c3f5-44b2-b677-acd23cdde73c",
		Type:           "accounts",
		Version:        0,
		Attributes: Attributes{
			Country:         "GB",
			BaseCurrency:    "GBP",
			AccountNumber:   "41426820",
			BankID:          "400300",
			BankIDCode:      "GBDSC",
			BIC:             "NWBKGB22",
			IBAN:            "GB11NWBK40030041426820",
			Title:           "Mr",
			FirstName:       "Francisco",
			BankAccountName: "Fernandez",
			AlternativeBankAccountNames: []string{
				"Paco Holder",
			},
			AccountClassification:   "Personal",
			JointAccount:            false,
			AccountMatchingOptOut:   false,
			SecondaryIdentification: "A1B2C3D5",
		},
	}

	accountThree = Account{
		ID:             "ad27e265-9605-4b4b-a0e5-3003ea9cc4de",
		OrganisationID: "eb0bd6f5-c3f5-44b2-b677-acd23cdde73c",
		Type:           "accounts",
		Version:        0,
		Attributes: Attributes{
			Country:         "GB",
			BaseCurrency:    "GBP",
			AccountNumber:   "41426821",
			BankID:          "400300",
			BankIDCode:      "GBDSC",
			BIC:             "NWBKGB22",
			IBAN:            "GB11NWBK40030041426821",
			Title:           "Ms",
			FirstName:       "Liza",
			BankAccountName: "Johnson",
			AlternativeBankAccountNames: []string{
				"Lize Holder",
			},
			AccountClassification:   "Personal",
			JointAccount:            false,
			AccountMatchingOptOut:   false,
			SecondaryIdentification: "A1B2C3D6",
		},
	}
)

const (
	// badInput is an 'evil' account ID value used to simulate in the mock test server a bad request.
	badInput = "bad-input"

	// badInputPageNumParamOpt is an 'evil' page number option parameter value used to simulate in the mock test server a bad request.
	badInputPageNumParamOpt = "bad-page-number"

	// badInputPageOpt is an 'evil' page option value used to simulate in the mock test server a bad request.
	badInputPageOpt = -1

	// serviceFailure is an 'evil' account ID value used to simulate in the mock test server a service failure.
	serviceFailure = "force-service-failure"

	// serviceFailurePageOpt is an 'evil' page option value used to simulate in the mock test server a service failure.
	serviceFailurePageOpt = -2
)

func TestCreate(t *testing.T) {
	type testData struct {
		existing []Account
		account  AccountResource
		want     *AccountResource
		err      error
	}

	var golds = []testData{
		0: {
			[]Account{},
			AccountResource{Data: accountOne},
			&AccountResource{Data: accountOne},
			nil,
		},
		1: {
			[]Account{accountOne},
			AccountResource{Data: accountOne},
			nil,
			ErrConflict,
		},
		2: {
			[]Account{},
			AccountResource{Data: createBadAccount()},
			nil,
			ErrBadInput,
		},
		3: {
			[]Account{},
			AccountResource{Data: createMonkeyAccount()},
			nil,
			ErrServerError,
		},
	}

	var testCreate = func(t *testing.T, tc int, data testData) {
		// Setup mocked account repository
		repo := setupAccountRepo(data.existing...)

		// Setup mocked server
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			assert.Equal(t, req.URL.String(), fmt.Sprintf("/v1/organisation/accounts"))

			var a AccountResource
			err := json.NewDecoder(req.Body).Decode(&a)
			assert.NoError(t, err)

			// Note: this is a very simplistic assumption of checking a hard pre-defined value
			// to simulate either a 'bad input' or a 'service failure' on the service side.
			if a.Data.ID == badInput {
				serveError(t, rw, http.StatusBadRequest)
			} else if a.Data.ID == serviceFailure {
				serveError(t, rw, http.StatusInternalServerError)
			} else if _, ok := repo[a.Data.ID]; ok {
				serveError(t, rw, http.StatusConflict)
			} else {
				serveContent(t, rw, http.StatusCreated, a)
			}
		}))
		defer server.Close()

		// Setup client
		client := setupClient(t, server.URL)

		got, err := client.Create(&data.account)

		assert.Equal(t, data.want, got, fmt.Sprintf("%d. Want account %+v, but got %+v", tc, data.want, got))
		assert.Equal(t, data.err, err, fmt.Sprintf("%d. Want error %+v, but got %+v", tc, data.err, err))
	}

	for i, g := range golds {
		testCreate(t, i, g)
	}
}

func TestFetch(t *testing.T) {
	type testData struct {
		existing  []Account
		accountID string
		want      *AccountResource
		err       error
	}

	var golds = []testData{
		0: {
			[]Account{accountOne},
			"ad27e265-9605-4b4b-a0e5-3003ea9cc4dc",
			&AccountResource{Data: accountOne},
			nil,
		},
		1: {
			[]Account{},
			"ad27e265-9605-4b4b-a0e5-3003ea9cc4dc",
			nil,
			ErrNotFound,
		},
		2: {
			[]Account{},
			badInput,
			nil,
			ErrBadInput,
		},
		3: {
			[]Account{},
			serviceFailure,
			nil,
			ErrServerError,
		},
	}

	var testFetch = func(t *testing.T, tc int, data testData) {
		// Setup mocked account repository
		repo := setupAccountRepo(data.existing...)

		// Setup mocked server
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			pathSegments := strings.Split(req.URL.Path, "/")
			ID := pathSegments[len(pathSegments)-1]

			assert.Equal(t, req.URL.String(), fmt.Sprintf("/v1/organisation/accounts/%s", ID))

			if ID == badInput {
				serveError(t, rw, http.StatusBadRequest)
			} else if ID == serviceFailure {
				serveError(t, rw, http.StatusInternalServerError)
			} else if a, ok := repo[ID]; ok {
				serveContent(t, rw, http.StatusOK, AccountResource{Data: a})
			} else {
				serveError(t, rw, http.StatusNotFound)
			}
		}))
		defer server.Close()

		// Setup client
		client := setupClient(t, server.URL)

		got, err := client.Fetch(data.accountID)

		assert.Equal(t, data.want, got, fmt.Sprintf("%d. Want account %+v, but got %+v", tc, data.want, got))
		assert.Equal(t, data.err, err, fmt.Sprintf("%d. Want error %+v, but got %+v", tc, data.err, err))
	}

	for i, g := range golds {
		testFetch(t, i, g)
	}
}

func TestList(t *testing.T) {
	type testData struct {
		existing []Account
		opts     PageOpts
		want     *AccountsResource
		err      error
	}

	var golds = []testData{
		0: {
			[]Account{},
			PageOpts{
				Number: PageNumOptOf("0"),
				Size:   PageSizeOptOf(100),
			},
			&AccountsResource{},
			nil,
		},
		1: {
			[]Account{accountOne},
			PageOpts{
				Size: PageSizeOptOf(100),
			},
			&AccountsResource{Data: []Account{accountOne}},
			nil,
		},
		2: {
			[]Account{accountOne},
			PageOpts{
				Number: PageNumOptOf("0"),
			},
			&AccountsResource{Data: []Account{accountOne}},
			nil,
		},
		3: {
			[]Account{accountOne},
			PageOpts{
				Number: PageNumOptOf("0"),
				Size:   PageSizeOptOf(100),
			},
			&AccountsResource{Data: []Account{accountOne}},
			nil,
		},
		4: {
			[]Account{accountOne, accountTwo, accountThree},
			PageOpts{
				Number: PageNumOptOf("1"),
				Size:   PageSizeOptOf(1),
			},
			&AccountsResource{Data: []Account{accountTwo}},
			nil,
		},
		5: {
			[]Account{accountOne, accountTwo, accountThree},
			PageOpts{
				Number: PageNumOptOf("0"),
				Size:   PageSizeOptOf(2),
			},
			&AccountsResource{Data: []Account{accountOne, accountTwo}},
			nil,
		},
		6: {
			[]Account{accountOne, accountTwo, accountThree},
			PageOpts{
				Number: PageNumOptOf("1"),
				Size:   PageSizeOptOf(2),
			},
			&AccountsResource{Data: []Account{accountThree}},
			nil,
		},
		7: {
			[]Account{accountOne, accountTwo, accountThree},
			PageOpts{
				Number: PageNumOptOf("1"),
				Size:   PageSizeOptOf(2),
			},
			&AccountsResource{Data: []Account{accountThree}},
			nil,
		},
		8: {
			[]Account{},
			PageOpts{
				Number: PageNumOptOf("first"),
				Size:   PageSizeOptOf(2),
			},
			&AccountsResource{},
			nil,
		},
		9: {
			[]Account{accountOne, accountTwo},
			PageOpts{
				Number: PageNumOptOf("first"),
				Size:   PageSizeOptOf(1),
			},
			&AccountsResource{Data: []Account{accountOne}},
			nil,
		},
		10: {
			[]Account{},
			PageOpts{
				Number: PageNumOptOf("last"),
				Size:   PageSizeOptOf(2),
			},
			&AccountsResource{},
			nil,
		},
		11: {
			[]Account{accountOne, accountTwo, accountThree},
			PageOpts{
				Number: PageNumOptOf("last"),
				Size:   PageSizeOptOf(1),
			},
			&AccountsResource{Data: []Account{accountThree}},
			nil,
		},
		12: {
			[]Account{accountOne, accountTwo, accountThree},
			PageOpts{
				Number: PageNumOptOf("last"),
				Size:   PageSizeOptOf(2),
			},
			&AccountsResource{Data: []Account{accountThree}},
			nil,
		},
		13: {
			[]Account{},
			PageOpts{
				Number: PageNumOptOf(badInputPageNumParamOpt),
				Size:   PageSizeOptOf(2),
			},
			nil,
			ErrBadInput,
		},
		14: {
			[]Account{},
			PageOpts{
				Number: PageNumOptOf("1"),
				Size:   PageSizeOptOf(serviceFailurePageOpt),
			},
			nil,
			ErrServerError,
		},
	}

	var testList = func(t *testing.T, tc int, data testData) {
		// Setup mocked account repository
		repo := setupAccountRepo(data.existing...)

		var pageOptsOf = func(req *http.Request) (int64, int64) {
			qryParams := req.URL.Query()
			assert.Equal(t, req.URL.String(), fmt.Sprintf("/v1/organisation/accounts?%s", qryParams.Encode()))

			// Default paging options
			pNum := int64(0)
			pSize := int64(100)
			var err error

			// Override paging options (handle the 'first' and 'last' use cases as well
			pSizeParam := qryParams.Get("page[size]")
			if pSizeParam != "" {
				pSize, err = strconv.ParseInt(pSizeParam, 10, 64)
				assert.NoError(t, err)
			}

			pNumParam := qryParams.Get("page[number]")
			if pNumParam == badInputPageNumParamOpt {
				pNum = badInputPageOpt
			} else if pNumParam == "last" {
				total := int64(len(repo))
				if total > 0 {
					if total%pSize == 0 {
						pNum = (total / pSize) - 1
					} else {
						pNum = total / pSize
					}
				}
			} else if pNumParam != "" && pNumParam != "first" {
				pNum, err = strconv.ParseInt(pNumParam, 10, 64)
				assert.NoError(t, err)
			}

			return pNum, pSize
		}

		var serveAccountsPage = func(t *testing.T, rw http.ResponseWriter, pNum, pSize int64) {
			// Sort the keys to rely on a sorted repository for the sake of the test
			total := int64(len(repo))
			start := pNum * pSize
			end := start + pSize
			var resultKeys []string
			if start < total {
				var keys []string
				for k := range repo {
					keys = append(keys, k)
				}
				sort.Strings(keys)

				if end >= total {
					// Partial page
					resultKeys = keys[start:]
				} else {
					// Total page
					resultKeys = keys[start:end]
				}
			}

			// Collect the accounts to be listed
			var accounts []Account
			for _, k := range resultKeys {
				if a, ok := repo[k]; ok {
					accounts = append(accounts, a)
				}
			}

			serveContent(t, rw, http.StatusOK, AccountsResource{Data: accounts})
		}

		// Setup mocked server
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			qryParams := req.URL.Query()
			assert.Equal(t, req.URL.String(), fmt.Sprintf("/v1/organisation/accounts?%s", qryParams.Encode()))

			// Resolve the effective page number and size
			pNum, pSize := pageOptsOf(req)

			if pNum == badInputPageOpt || pSize == badInputPageOpt {
				serveError(t, rw, http.StatusBadRequest)
			} else if pNum == serviceFailurePageOpt || pSize == serviceFailurePageOpt {
				serveError(t, rw, http.StatusInternalServerError)
			} else {
				serveAccountsPage(t, rw, pNum, pSize)
			}
		}))
		defer server.Close()

		// Setup client
		client := setupClient(t, server.URL)

		got, err := client.List(&data.opts)

		assert.Equal(t, data.want, got, fmt.Sprintf("%d. Want accounts %+v, but got %+v", tc, data.want, got))
		assert.Equal(t, data.err, err, fmt.Sprintf("%d. Want error %+v, but got %+v", tc, data.err, err))
	}

	for i, g := range golds {
		testList(t, i, g)
	}
}

func TestDelete(t *testing.T) {
	type testData struct {
		existing  []Account
		accountID string
		version   int64
		err       error
	}

	var golds = []testData{
		0: {
			[]Account{accountOne},
			accountOne.ID,
			accountOne.Version,
			nil,
		},
		1: {
			[]Account{accountOne},
			accountOne.ID,
			100,
			ErrConflict,
		},
		2: {
			[]Account{},
			accountOne.ID,
			accountOne.Version,
			ErrNotFound,
		},
	}

	var testDelete = func(t *testing.T, tc int, data testData) {
		// Setup mocked account repository
		repo := setupAccountRepo(data.existing...)

		// Setup mocked server
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			pathSegments := strings.Split(req.URL.Path, "/")
			ID := pathSegments[len(pathSegments)-1]
			qryParams := req.URL.Query()
			version, err := strconv.ParseInt(qryParams.Get("version"), 10, 64)
			assert.NoError(t, err)

			assert.Equal(t, fmt.Sprintf("/v1/organisation/accounts/%s?version=%d", ID, version), req.URL.String())

			if ID == badInput {
				serveError(t, rw, http.StatusBadRequest)
			} else if ID == serviceFailure {
				serveError(t, rw, http.StatusInternalServerError)
			} else if a, ok := repo[ID]; ok {
				if a.Version == version {
					rw.WriteHeader(http.StatusNoContent)
				} else {
					rw.WriteHeader(http.StatusConflict)
				}
			} else {
				serveError(t, rw, http.StatusNotFound)
			}
		}))
		defer server.Close()

		// Setup client
		client := setupClient(t, server.URL)

		err := client.Delete(data.accountID, data.version)

		assert.Equal(t, data.err, err, fmt.Sprintf("%d. Want error %+v, but got %+v", tc, data.err, err))
	}

	for i, g := range golds {
		testDelete(t, i, g)
	}
}

func setupClient(t *testing.T, baseURL string) *Client {
	u, err := url.Parse(baseURL)
	if err != nil {
		t.Fatalf("failed to setup client. %s", err)
	}
	return &Client{
		BaseURL: u,
	}
}

func setupAccountRepo(accounts ...Account) map[string]Account {
	r := make(map[string]Account, len(accounts))

	for _, a := range accounts {
		r[a.ID] = a
	}

	return r
}

func serveError(t *testing.T, rw http.ResponseWriter, statusCode int) {
	errDetail, err := json.Marshal(
		errorDetail{
			ErrorCode: fmt.Sprintf("%d", statusCode),
			ErrorMsg:  http.StatusText(statusCode),
		})
	assert.NoError(t, err)

	rw.WriteHeader(statusCode)
	_, err = rw.Write(errDetail)
	assert.NoError(t, err)
}

func serveContent(t *testing.T, rw http.ResponseWriter, statusCode int, content interface{}) {
	body, err := json.Marshal(content)
	assert.NoError(t, err)

	rw.WriteHeader(statusCode)
	_, err = rw.Write(body)
	assert.NoError(t, err)
}

// createBadAccount creates a simulation account which will be considered a bad input on the server side.
func createBadAccount() Account {
	return Account{
		ID: badInput,
	}
}

// createMonkeyAccount creates a simulation account which will cause an unknown service failure on the server side.
func createMonkeyAccount() Account {
	return Account{
		ID: serviceFailure,
	}
}
