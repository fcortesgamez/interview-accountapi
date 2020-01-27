# Francisco Cortes Gamez

Solution to the Form3 Account API Take Home Exercise

## How to run the tests

* `docker-compose up --build --abort-on-container-exit` to run the unit and _Pact-based_ integration (contract) tests.

## Alternative using `make`
 
* `make unit` to run the unit tests.
* `make integration` to run the _Pact-based_ integration (contract) tests.

Then, in case you wish to publish the _Pact_ generated after running the integration tests:

1. `docker-compose up -d pactbroker` to run the _Pact Broker_.
1. `make publish` to publish the generated pacts to the _Pact broker_ running in localhost.
1. In your browser, go to `http://localhost:8081/pacts/provider/AccountApi/consumer/AccountApiClient/latest` to see the published pacts.
1. Login using _username_ `pact_accountapi`, password `pact_accountapi`.

Then you should see your generated _pacts_ like in the screen below.

![Pacts in your Pact Broker](/docs/images/pact-interactions.png "Pacts in your Pact Broker")

## Tiny example app against the provided Accounts API

It is possible to run an example app against the provided Accounts API.

1. `docker-compose up -d accountapi` to run the _Accounts API_ service.
1. `make run-client` to run the example app.

## Project structure

* Folder `client` contains the client code, unit and _Pact based_ tests.
* Folder `client/cmd` contains a simple example app to run against the provided Accounts API. 
* Folder `client/pact` contains a simple app used to publish the _pacts_ to the _Pacts Broker_.

## Remarks

* A conscious decision has been taken to implement the integration tests as contract tests using _Pact_.
* Please visit [pact.io](https://docs.pact.io/) to find out about the pros and cons of using _Pact_ for your integration tests.  
* Please visit [pact.io - Implementation guidelines: Go](https://docs.pact.io/implementation_guides/go) for a reference on how to implement _Pact based_ tests in _Go_ applications. 

