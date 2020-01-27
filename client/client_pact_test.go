// +build integration

package client

import (
	"errors"
	"fmt"
	"github.com/pact-foundation/pact-go/dsl"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"
)

var (
	pact dsl.Pact

	account = AccountResource{
		Data: Account{
			ID:             ID,
			OrganisationID: "eb0bd6f5-c3f5-44b2-b677-acd23cdde73c",
			Type:           "accounts",
			Version:        0,
			Attributes: Attributes{
				Country:                 "GB",
				BaseCurrency:            "GBP",
				AccountNumber:           "41426819",
				BankID:                  "400300",
				BankIDCode:              "GBDSC",
				BIC:                     "NWBKGB22",
				IBAN:                    "GB11NWBK40030041426819",
				Title:                   "Ms",
				FirstName:               "Samantha",
				BankAccountName:         "Samantha Holder",
				AccountClassification:   "Personal",
				JointAccount:            false,
				AccountMatchingOptOut:   false,
				SecondaryIdentification: "A1B2C3D4",
			},
		},
	}

	badAccount = AccountResource{
		Data: Account{
			ID:             notAGuiID,
			OrganisationID: "eb0bd6f5-c3f5-44b2-b677-acd23cdde73c",
			Type:           "accounts",
			Version:        0,
			Attributes: Attributes{
				Country:                 "GB",
				BaseCurrency:            "GBP",
				AccountNumber:           "41426819",
				BankID:                  "400300",
				BankIDCode:              "GBDSC",
				BIC:                     "NWBKGB22",
				IBAN:                    "GB11NWBK40030041426819",
				Title:                   "Ms",
				FirstName:               "Samantha",
				BankAccountName:         "Samantha Holder",
				AccountClassification:   "Personal",
				JointAccount:            false,
				AccountMatchingOptOut:   false,
				SecondaryIdentification: "A1B2C3D4",
			},
		},
	}
)

const (
	ID        = "ad27e265-9605-4b4b-a0e5-3003ea9cc4dc"
	notAGuiID = "not-a-gui-id"

	// States
	someAccountsPresent   = "some accounts exist"
	accountPresent        = "the account already exists"
	accountNotPresent     = "the account does not exists"
	serviceUnknownFailure = "an unknown service failure"
)

var u *url.URL
var client *Client

func TestMain(m *testing.M) {
	var exitCode int

	// Setup Pact and related test stuff
	setup()

	// Run all the tests
	exitCode = m.Run()

	// Shutdown the Mock Service and Write pact files to disk
	if err := pact.WritePact(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	pact.Teardown()
	os.Exit(exitCode)
}

func TestPactCreate(t *testing.T) {
	req := dsl.Request{
		Method: http.MethodPost,
		Path:   dsl.String("/v1/organisation/accounts"),
		Body:   dsl.Like(account),
		Headers: dsl.MapMatcher{
			"Content-Type": dsl.String("application/vnd.api+json"),
			"Accept":       dsl.String("application/vnd.api+json"),
			"User-Agent":   dsl.String("Accounts API Go client"),
		},
	}

	t.Run("create an account which does not exist", func(t *testing.T) {
		pact.
			AddInteraction().
			Given(accountNotPresent).
			UponReceiving("a request to create an account").
			WithRequest(req).
			WillRespondWith(dsl.Response{
				Status: http.StatusCreated,
				Body:   dsl.Like(account),
				Headers: dsl.MapMatcher{
					"Content-Type": dsl.String("application/vnd.api+json"),
				},
			})

		err := pact.Verify(func() error {
			a, err := client.Create(&AccountResource{})
			if a == nil {
				return errors.New("no account returned, but expected")
			}

			return err
		})

		if err != nil {
			t.Fatalf("verification error: %v", err)
		}
	})

	t.Run("does not create an account which exists already", func(t *testing.T) {
		pact.
			AddInteraction().
			Given(accountPresent).
			UponReceiving("a request to create an account").
			WithRequest(req).
			WillRespondWith(dsl.Response{
				Status: http.StatusConflict,
				Body:   dsl.Like(account),
				Headers: dsl.MapMatcher{
					"Content-Type": dsl.String("application/vnd.api+json"),
				},
			})

		err := pact.Verify(func() error {
			a, err := client.Create(&AccountResource{})
			if a != nil {
				return errors.New("account returned, but not expected")
			}
			if err != ErrConflict {
				return fmt.Errorf("error %s, but expected %s", err, ErrConflict)
			}
			return nil
		})

		if err != nil {
			t.Fatalf("verification error: %v", err)
		}
	})

	t.Run("does not create an account in case of a bad request", func(t *testing.T) {
		pact.
			AddInteraction().
			Given(accountNotPresent).
			UponReceiving("a bad request to create an account").
			WithRequest(
				dsl.Request{
					Method: http.MethodPost,
					Path:   dsl.String("/v1/organisation/accounts"),
					Body:   dsl.Like(badAccount),
					Headers: dsl.MapMatcher{
						"Content-Type": dsl.String("application/vnd.api+json"),
						"Accept":       dsl.String("application/vnd.api+json"),
						"User-Agent":   dsl.String("Accounts API Go client"),
					},
				}).
			WillRespondWith(dsl.Response{
				Status: http.StatusBadRequest,
				Body:   dsl.Like(account),
				Headers: dsl.MapMatcher{
					"Content-Type": dsl.String("application/vnd.api+json"),
				},
			})

		err := pact.Verify(func() error {
			a, err := client.Create(&AccountResource{})
			if a != nil {
				return errors.New("account returned, but not expected")
			}
			if err != ErrBadInput {
				return fmt.Errorf("error %s, but expected %s", err, ErrBadInput)
			}
			return nil
		})

		if err != nil {
			t.Fatalf("verification error: %v", err)
		}
	})

	t.Run("does not create an account in case of unknown service failure", func(t *testing.T) {
		pact.
			AddInteraction().
			Given(serviceUnknownFailure).
			UponReceiving("a request to create an account").
			WithRequest(req).
			WillRespondWith(dsl.Response{
				Status: http.StatusInternalServerError,
				Body:   dsl.Like(account),
				Headers: dsl.MapMatcher{
					"Content-Type": dsl.String("application/vnd.api+json"),
				},
			})

		err := pact.Verify(func() error {
			a, err := client.Create(&AccountResource{})
			if a != nil {
				return errors.New("account returned, but not expected")
			}
			if err != ErrServerError {
				return fmt.Errorf("error %s, but expected %s", err, ErrServerError)
			}
			return nil
		})

		if err != nil {
			t.Fatalf("verification error: %v", err)
		}
	})
}

func TestPactFetch(t *testing.T) {
	req := dsl.Request{
		Method: http.MethodGet,
		Path:   dsl.String("/v1/organisation/accounts/ad27e265-9605-4b4b-a0e5-3003ea9cc4dc"),
		Headers: dsl.MapMatcher{
			"Accept":     dsl.String("application/vnd.api+json"),
			"User-Agent": dsl.String("Accounts API Go client"),
		},
	}

	t.Run("fetch an existing account", func(t *testing.T) {
		pact.
			AddInteraction().
			Given(accountPresent).
			UponReceiving("a request to fetch an account").
			WithRequest(req).
			WillRespondWith(dsl.Response{
				Status: http.StatusOK,
				Body:   dsl.Match(AccountResource{}),
				Headers: dsl.MapMatcher{
					"Content-Type": dsl.String("application/vnd.api+json"),
				},
			})

		err := pact.Verify(func() error {
			a, err := client.Fetch(ID)

			if a == nil {
				return errors.New("no account fetched, but expected")
			}
			return err
		})

		if err != nil {
			t.Fatalf("verification error: %v", err)
		}
	})

	t.Run("does not fetch a non-existing account", func(t *testing.T) {
		pact.
			AddInteraction().
			Given(accountNotPresent).
			UponReceiving("a request to fetch an account").
			WithRequest(req).
			WillRespondWith(dsl.Response{
				Status: http.StatusNotFound,
				Body:   dsl.Match(AccountResource{}),
				Headers: dsl.MapMatcher{
					"Content-Type": dsl.String("application/vnd.api+json"),
				},
			})

		err := pact.Verify(func() error {
			a, err := client.Fetch(ID)

			if a != nil {
				return errors.New("account fetched, but not expected")
			}
			if err != ErrNotFound {
				return fmt.Errorf("error %s, but expected %s", err, ErrNotFound)
			}
			return nil
		})

		if err != nil {
			t.Fatalf("verification error: %v", err)
		}
	})

	t.Run("does not fetch any account in case of a bad request", func(t *testing.T) {
		pact.
			AddInteraction().
			Given(accountPresent).
			UponReceiving("a bad request to fetch an account").
			WithRequest(dsl.Request{
				Method: http.MethodGet,
				Path:   dsl.String(fmt.Sprintf("/v1/organisation/accounts/%s", notAGuiID)),
				Headers: dsl.MapMatcher{
					"Accept":     dsl.String("application/vnd.api+json"),
					"User-Agent": dsl.String("Accounts API Go client"),
				},
			}).
			WillRespondWith(dsl.Response{
				Status: http.StatusBadRequest,
				Body:   dsl.Match(AccountResource{}),
				Headers: dsl.MapMatcher{
					"Content-Type": dsl.String("application/vnd.api+json"),
				},
			})

		err := pact.Verify(func() error {
			a, err := client.Fetch(notAGuiID)

			if a != nil {
				return errors.New("account fetched, but not expected")
			}
			if err != ErrBadInput {
				return fmt.Errorf("error %s, but expected %s", err, ErrBadInput)
			}
			return nil
		})

		if err != nil {
			t.Fatalf("verification error: %v", err)
		}
	})

	t.Run("does not fetch any account in case of unknown service failure", func(t *testing.T) {
		pact.
			AddInteraction().
			Given(serviceUnknownFailure).
			UponReceiving("a request to fetch an account").
			WithRequest(req).
			WillRespondWith(dsl.Response{
				Status: http.StatusInternalServerError,
				Body:   dsl.Match(AccountResource{}),
				Headers: dsl.MapMatcher{
					"Content-Type": dsl.String("application/vnd.api+json"),
				},
			})

		err := pact.Verify(func() error {
			a, err := client.Fetch(ID)

			if a != nil {
				return errors.New("account fetched, but not expected")
			}
			if err != ErrServerError {
				return fmt.Errorf("error %s, but expected %s", err, ErrServerError)
			}
			return nil
		})

		if err != nil {
			t.Fatalf("verification error: %v", err)
		}
	})
}

func TestPactList(t *testing.T) {
	t.Run("list the existing accounts using default paging", func(t *testing.T) {
		pact.
			AddInteraction().
			Given(someAccountsPresent).
			UponReceiving("a request to list the existing accounts using default paging").
			WithRequest(dsl.Request{
				Method: http.MethodGet,
				Path:   dsl.String("/v1/organisation/accounts"),
				Headers: dsl.MapMatcher{
					"Accept":     dsl.String("application/vnd.api+json"),
					"User-Agent": dsl.String("Accounts API Go client"),
				},
			}).
			WillRespondWith(dsl.Response{
				Status: http.StatusOK,
				Body:   dsl.Match(AccountsResource{}),
				Headers: dsl.MapMatcher{
					"Content-Type": dsl.String("application/vnd.api+json"),
				},
			})

		err := pact.Verify(func() error {
			accounts, err := client.List(&PageOpts{})

			if accounts == nil {
				return errors.New("no accounts listed, but expected")
			}
			return err
		})

		if err != nil {
			t.Fatalf("verification error: %v", err)
		}
	})

	t.Run("list the existing accounts using custom paging", func(t *testing.T) {
		pact.
			AddInteraction().
			Given(someAccountsPresent).
			UponReceiving("a request to list the existing accounts using custom paging").
			WithRequest(dsl.Request{
				Method: http.MethodGet,
				Path:   dsl.String("/v1/organisation/accounts"),
				Query: dsl.MapMatcher{
					"page[number]": dsl.Term("2", "[0-9]+"),
					"page[size]":   dsl.Term("50", "[0-9]+"),
				},
				Headers: dsl.MapMatcher{
					"Accept":     dsl.String("application/vnd.api+json"),
					"User-Agent": dsl.String("Accounts API Go client"),
				},
			}).
			WillRespondWith(dsl.Response{
				Status: http.StatusOK,
				Body:   dsl.Match(AccountsResource{}),
				Headers: dsl.MapMatcher{
					"Content-Type": dsl.String("application/vnd.api+json"),
				},
			})

		err := pact.Verify(func() error {
			accounts, err := client.List(&PageOpts{
				Number: PageNumOptOf("2"),
				Size:   PageSizeOptOf(50),
			})

			if accounts == nil {
				return errors.New("no accounts listed, but expected")
			}
			return err
		})

		if err != nil {
			t.Fatalf("verification error: %v", err)
		}
	})

	t.Run("does not list any account in case of a bad request", func(t *testing.T) {
		pact.
			AddInteraction().
			Given(someAccountsPresent).
			UponReceiving("a bad request to list the existing accounts").
			WithRequest(dsl.Request{
				Method: http.MethodGet,
				Path:   dsl.String("/v1/organisation/accounts"),
				Query: dsl.MapMatcher{
					"page[number]": dsl.Term("2", "[0-9]+"),
					"page[size]":   dsl.Term("50", "[0-9]+"),
				},
				Headers: dsl.MapMatcher{
					"Accept":     dsl.String("application/vnd.api+json"),
					"User-Agent": dsl.String("Accounts API Go client"),
				},
			}).
			WillRespondWith(dsl.Response{
				Status: http.StatusBadRequest,
				Body:   dsl.Match(AccountsResource{}),
				Headers: dsl.MapMatcher{
					"Content-Type": dsl.String("application/vnd.api+json"),
				},
			})

		err := pact.Verify(func() error {
			accounts, err := client.List(&PageOpts{
				Number: PageNumOptOf("2"),
				Size:   PageSizeOptOf(50),
			})

			if accounts != nil {
				return errors.New("accounts listed, but not expected")
			}
			if err != ErrBadInput {
				return fmt.Errorf("error %s, but expected %s", err, ErrBadInput)
			}
			return nil
		})

		if err != nil {
			t.Fatalf("verification error: %v", err)
		}
	})

	t.Run("does not list any account in case of unknown service failure", func(t *testing.T) {
		pact.
			AddInteraction().
			Given(serviceUnknownFailure).
			UponReceiving("a request to list the existing accounts").
			WithRequest(dsl.Request{
				Method: http.MethodGet,
				Path:   dsl.String("/v1/organisation/accounts"),
				Query: dsl.MapMatcher{
					"page[number]": dsl.Term("2", "[0-9]+"),
					"page[size]":   dsl.Term("50", "[0-9]+"),
				},
				Headers: dsl.MapMatcher{
					"Accept":     dsl.String("application/vnd.api+json"),
					"User-Agent": dsl.String("Accounts API Go client"),
				},
			}).
			WillRespondWith(dsl.Response{
				Status: http.StatusInternalServerError,
				Body:   dsl.Match(AccountsResource{}),
				Headers: dsl.MapMatcher{
					"Content-Type": dsl.String("application/vnd.api+json"),
				},
			})

		err := pact.Verify(func() error {
			accounts, err := client.List(&PageOpts{
				Number: PageNumOptOf("2"),
				Size:   PageSizeOptOf(50),
			})

			if accounts != nil {
				return errors.New("accounts listed, but not expected")
			}
			if err != ErrServerError {
				return fmt.Errorf("error %s, but expected %s", err, ErrServerError)
			}
			return nil
		})

		if err != nil {
			t.Fatalf("verification error: %v", err)
		}
	})
}

func TestPactDelete(t *testing.T) {
	req := dsl.Request{
		Method: http.MethodDelete,
		Path:   dsl.String("/v1/organisation/accounts/ad27e265-9605-4b4b-a0e5-3003ea9cc4dc"),
		Query: dsl.MapMatcher{
			"version": dsl.Term("0", "[0-9]+"),
		},
		Headers: dsl.MapMatcher{
			"Accept":     dsl.String("application/vnd.api+json"),
			"User-Agent": dsl.String("Accounts API Go client"),
		},
	}

	t.Run("delete an existing account", func(t *testing.T) {
		pact.
			AddInteraction().
			Given(accountPresent).
			UponReceiving("a request to delete an account").
			WithRequest(req).
			WillRespondWith(dsl.Response{
				Status: http.StatusNoContent,
				//Body:   dsl.Match(Account{}),
				Headers: dsl.MapMatcher{
					"Content-Type": dsl.String("application/vnd.api+json"),
				},
			})

		err := pact.Verify(func() error {
			return client.Delete(ID, 0)
		})

		if err != nil {
			t.Fatalf("verification error: %v", err)
		}
	})

	t.Run("does not delete an existing account using a non-current version", func(t *testing.T) {
		pact.
			AddInteraction().
			Given(accountPresent).
			UponReceiving("a request to delete an account using a non-current version").
			WithRequest(req).
			WillRespondWith(dsl.Response{
				Status: http.StatusConflict,
				Headers: dsl.MapMatcher{
					"Content-Type": dsl.String("application/vnd.api+json"),
				},
			})

		err := pact.Verify(func() error {
			err := client.Delete(ID, 0)

			if err != ErrConflict {
				return fmt.Errorf("error %s, but expected %s", err, ErrConflict)
			}

			return nil
		})

		if err != nil {
			t.Fatalf("verification error: %v", err)
		}
	})

	t.Run("does not delete a non-existing account", func(t *testing.T) {
		pact.
			AddInteraction().
			Given(accountNotPresent).
			UponReceiving("a request to delete an account").
			WithRequest(req).
			WillRespondWith(dsl.Response{
				Status: http.StatusNotFound,
				Headers: dsl.MapMatcher{
					"Content-Type": dsl.String("application/vnd.api+json"),
				},
			})

		err := pact.Verify(func() error {
			err := client.Delete(ID, 0)

			if err != ErrNotFound {
				return fmt.Errorf("error %s, but expected %s", err, ErrNotFound)
			}

			return nil
		})

		if err != nil {
			t.Fatalf("verification error: %v", err)
		}
	})

	t.Run("does not delete any account in case of a bad request", func(t *testing.T) {
		pact.
			AddInteraction().
			Given(accountPresent).
			UponReceiving("a bad request to delete an account").
			WithRequest(dsl.Request{
				Method: http.MethodDelete,
				Path:   dsl.String(fmt.Sprintf("/v1/organisation/accounts/%s", notAGuiID)),
				Query: dsl.MapMatcher{
					"version": dsl.Term("0", "[0-9]+"),
				},
				Headers: dsl.MapMatcher{
					"Accept":     dsl.String("application/vnd.api+json"),
					"User-Agent": dsl.String("Accounts API Go client"),
				},
			}).
			WillRespondWith(dsl.Response{
				Status: http.StatusBadRequest,
				Headers: dsl.MapMatcher{
					"Content-Type": dsl.String("application/vnd.api+json"),
				},
			})

		err := pact.Verify(func() error {
			err := client.Delete(notAGuiID, 0)

			if err != ErrBadInput {
				return fmt.Errorf("error %s, but expected %s", err, ErrBadInput)
			}

			return nil
		})

		if err != nil {
			t.Fatalf("verification error: %v", err)
		}
	})

	t.Run("does not delete any account in case of unknown service failure", func(t *testing.T) {
		pact.
			AddInteraction().
			Given(serviceUnknownFailure).
			UponReceiving("a request to delete an account").
			WithRequest(req).
			WillRespondWith(dsl.Response{
				Status: http.StatusInternalServerError,
				Headers: dsl.MapMatcher{
					"Content-Type": dsl.String("application/vnd.api+json"),
				},
			})

		err := pact.Verify(func() error {
			err := client.Delete(ID, 0)

			if err != ErrServerError {
				return fmt.Errorf("error %s, but expected %s", err, ErrServerError)
			}

			return nil
		})

		if err != nil {
			t.Fatalf("verification error: %v", err)
		}
	})
}

func setup() {
	pact = dsl.Pact{
		Consumer:                 os.Getenv("CONSUMER_NAME"),
		Provider:                 os.Getenv("PROVIDER_NAME"),
		LogDir:                   os.Getenv("LOG_DIR"),
		PactDir:                  os.Getenv("PACT_DIR"),
		LogLevel:                 "INFO",
		DisableToolValidityCheck: true,
		ClientTimeout:            time.Duration(120) * time.Second,
	}

	// Proactively start service to get access to the port
	pact.Setup(true)

	u, _ = url.Parse(fmt.Sprintf("http://localhost:%d", pact.Server.Port))

	client = &Client{
		BaseURL: u,
	}
}
