export PATH := $(PWD)/pact/bin:$(PATH)
export PATH
export PROVIDER_NAME = AccountApi
export CONSUMER_NAME = AccountApiClient
export PACT_DIR = $(PWD)/pacts
export LOG_DIR = $(PWD)/log
export PACT_BROKER_PROTO = http
export PACT_BROKER_URL = localhost:8081
export PACT_BROKER_USERNAME = pact_accountapi
export PACT_BROKER_PASSWORD = pact_accountapi
