#!/bin/bash

echo "--- Installing Pact CLI dependencies"
curl -fsSL https://raw.githubusercontent.com/pact-foundation/pact-ruby-standalone/master/install.sh | bash

echo "--- Running tests"
make unit integration
