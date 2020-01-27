package main

import (
	"encoding/json"
	client2 "gitlab.com/kitolabs-private/form3/interview-accountapi/client"
	"log"
	"net/url"
)

// Small application to show the client running against the Accounts API deployed in the docker-compose containers.
func main() {
	c := setupClient("http://localhost:8080")

	var a client2.AccountResource
	sampleAccount(&a)

	create(c, &a)
	fetch(c, a.Data.ID)
	list(c)
	delete(c, a.Data.ID)
}

func create(c *client2.Client, a *client2.AccountResource) {
	created, err := c.Create(a)
	if err != nil {
		log.Printf("failed to create account %s", err)
	}
	if created != nil {
		log.Printf("created %+v", prettyPrintedJSON(created))
	}
}

func fetch(c *client2.Client, ID string) {
	a, err := c.Fetch(ID)
	if err != nil {
		log.Fatalf("failed to fetch account %+v, %s", a, err)
	}
	log.Printf("fetched: %+v\n", prettyPrintedJSON(a))
}

func list(c *client2.Client) {
	opts := client2.PageOpts{
		Number: client2.PageNumOptOf("0"),
		Size:   client2.PageSizeOptOf(0),
	}
	accounts, err := c.List(&opts)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("list: %+v\n", prettyPrintedJSON(accounts))
}

func delete(c *client2.Client, accountID string) {
	err := c.Delete(accountID, 0)
	if err != nil {
		log.Fatal(err)
	}
}

func setupClient(baseUrl string) *client2.Client {
	u, _ := url.Parse(baseUrl)
	return &client2.Client{
		BaseURL: u,
	}
}

func sampleAccount(account *client2.AccountResource) {
	*account = client2.AccountResource{
		Data: client2.Account{
			ID:             "ad27e265-9605-4b4b-a0e5-3003ea9cc4dc",
			OrganisationID: "eb0bd6f5-c3f5-44b2-b677-acd23cdde73c",
			Type:           "accounts",
			Version:        0,
			Attributes: client2.Attributes{
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
		},
	}
}

func prettyPrintedJSON(v interface{}) string {
	bytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Fatalf("failed to marshall with pretty-print option. %s", err)
	}
	return string(bytes)
}
