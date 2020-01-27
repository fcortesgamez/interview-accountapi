package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/go-querystring/query"
	"io"
	"log"
	"net/http"
	"net/url"
)

// Client is our consumer interface to the Accounts API.
type Client struct {
	BaseURL    *url.URL
	httpClient *http.Client
}

// An AccountsResource is a wrapper around multiple Account values used for serialization.
type AccountsResource struct {
	Data []Account `json:"data"`
}

// An AccountResource is a wrapper around a single Account value used for serialization.
type AccountResource struct {
	Data Account `json:"data"`
}

// Account represents a bank account that is registered with Form3.
type Account struct {
	ID             string     `json:"id"`
	OrganisationID string     `json:"organisation_id"`
	Type           string     `json:"type"`
	Version        int64      `json:"version"`
	Attributes     Attributes `json:"attributes"`
}

// Attributes represent the account attributes as per the Form3 specifications.
type Attributes struct {
	Country                     string   `json:"country"`
	BaseCurrency                string   `json:"base_currency"`
	BankID                      string   `json:"bank_id"`
	BankIDCode                  string   `json:"bank_id_code"`
	AccountNumber               string   `json:"account_number"`
	BIC                         string   `json:"bic"`
	IBAN                        string   `json:"iban"`
	CustomerID                  string   `json:"customer_id"`
	Title                       string   `json:"title"`
	FirstName                   string   `json:"first_name"`
	BankAccountName             string   `json:"bank_account_name"`
	AlternativeBankAccountNames []string `json:"alternative_bank_account_names"`
	AccountClassification       string   `json:"account_classification"`
	JointAccount                bool     `json:"joint_account"`
	AccountMatchingOptOut       bool     `json:"account_matching_opt_out"`
	SecondaryIdentification     string   `json:"secondary_identification"`
}

// errorDetail represents an error in the response body returned by the service.
type errorDetail struct {
	ErrorCode string `json:"error_code"`
	ErrorMsg  string `json:"error_message"`
}

// PageOpts represents some paging options. Use it in the method List() to specify custom paging.
type PageOpts struct {
	Number PageNumOpt  `url:"page[number],omitempty"`
	Size   PageSizeOpt `url:"page[size],omitempty"`
}

// PageNumOpt represents an optional page number. It should hold numeric values or the special values 'first' or 'last'
type PageNumOpt *string

// PageSizeOpt represents an optional page size.
type PageSizeOpt *int64

var (
	// Errors
	ErrNotFound    = errors.New("not found")
	ErrConflict    = errors.New("conflict")
	ErrBadInput    = errors.New("bad input")
	ErrServerError = errors.New("internal server error")
	ErrUnknown     = errors.New("unknown")
)

// PageNumOptOf returns the typed PageNumOpt for the given value v.
func PageNumOptOf(v string) PageNumOpt {
	return &v
}

// PageSizeOptOf returns the typed PageSizeOpt for the given value v.
func PageSizeOptOf(v int64) PageSizeOpt {
	return &v
}

// Create register the given bank account with Form3 or create a new one.
func (c *Client) Create(account *AccountResource) (*AccountResource, error) {
	path := fmt.Sprintf("/v1/organisation/accounts")

	req, err := c.newRequest(http.MethodPost, path, "", account)
	if err != nil {
		return nil, err
	}

	var created AccountResource
	resp, err := c.do(req, &created)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusCreated:
		return &created, nil
	case http.StatusConflict:
		return nil, ErrConflict
	default:
		if resp.StatusCode >= http.StatusInternalServerError {
			logError(resp)
			return nil, ErrServerError
		} else if resp.StatusCode >= http.StatusBadRequest {
			logError(resp)
			return nil, ErrBadInput
		}
		// Unknown error (not according the specifications)
		logError(resp)
		return nil, ErrUnknown
	}
}

// Fetch fetches the account referenced by the given accountID.
func (c *Client) Fetch(accountID string) (*AccountResource, error) {
	path := fmt.Sprintf("/v1/organisation/accounts/%s", accountID)

	req, err := c.newRequest(http.MethodGet, path, "", nil)
	if err != nil {
		return nil, err
	}

	var account AccountResource
	resp, err := c.do(req, &account)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return &account, nil
	case http.StatusNotFound:
		return nil, ErrNotFound
	default:
		if resp.StatusCode >= http.StatusInternalServerError {
			logError(resp)
			return nil, ErrServerError
		} else if resp.StatusCode >= http.StatusBadRequest {
			logError(resp)
			return nil, ErrBadInput
		}
		// Unknown error (not according the specifications)
		logError(resp)
		return nil, ErrUnknown
	}
}

// List lists all the presents accounts using the given paging options opts.
func (c *Client) List(opts *PageOpts) (*AccountsResource, error) {

	// Resolve the query string (only paging support)
	qryString := ""
	qryParams, _ := query.Values(opts)
	if len(qryParams) > 0 {
		qryString = qryParams.Encode()
	}

	req, err := c.newRequest(http.MethodGet, "/v1/organisation/accounts", qryString, nil)
	if err != nil {
		return nil, err
	}

	var accounts AccountsResource
	resp, err := c.do(req, &accounts)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return &accounts, nil
	default:
		if resp.StatusCode >= http.StatusInternalServerError {
			logError(resp)
			return nil, ErrServerError
		} else if resp.StatusCode >= http.StatusBadRequest {
			logError(resp)
			return nil, ErrBadInput
		}
		// Unknown error (not according the specifications)
		logError(resp)
		return nil, ErrUnknown
	}
}

// Delete deletes an account referenced by the given accountID and version.
func (c *Client) Delete(accountID string, version int64) error {
	path := fmt.Sprintf("/v1/organisation/accounts/%s", accountID)
	qryString := fmt.Sprintf("version=%d", version)

	req, err := c.newRequest(http.MethodDelete, path, qryString, nil)
	if err != nil {
		return err
	}

	resp, err := c.do(req, nil)
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case http.StatusNoContent:
		return nil
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusConflict:
		return ErrConflict
	default:
		if resp.StatusCode >= http.StatusInternalServerError {
			logError(resp)
			return ErrServerError
		} else if resp.StatusCode >= http.StatusBadRequest {
			logError(resp)
			return ErrBadInput
		}
		// Unknown error (not according the specifications)
		logError(resp)
		return ErrUnknown
	}
}

func (c *Client) newRequest(method, path, qryString string, body interface{}) (*http.Request, error) {
	rel := url.URL{Path: path}
	if qryString != "" {
		rel.RawQuery = qryString
	}
	u := c.BaseURL.ResolveReference(&rel)

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/vnd.api+json")
	}

	req.Header.Set("Accept", "application/vnd.api+json")
	req.Header.Set("User-Agent", "Accounts API Go client")

	return req, nil
}

func (c *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	if c.httpClient == nil {
		c.httpClient = http.DefaultClient
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if v != nil {
		err = json.NewDecoder(resp.Body).Decode(&v)
		if err != nil {
			return nil, err
		}
	}

	err = resp.Body.Close()
	if err != nil {
		return resp, err
	}

	return resp, nil
}

func logError(resp *http.Response) {
	var errDetail errorDetail

	err := json.NewDecoder(resp.Body).Decode(&errDetail)
	if err != nil {
		log.Printf("failed to decode error response. request: %s. error: %s\n", resp.Request.URL, err)
	}

	log.Printf("failed to call Account api. request: %s. error: %s %s\n", resp.Request.URL.String(), errDetail.ErrorCode, errDetail.ErrorMsg)
}
