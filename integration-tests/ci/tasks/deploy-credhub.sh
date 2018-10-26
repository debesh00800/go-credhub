#!/bin/bash 

set -eu

BASEDIR=$(pwd)

mv bbl-cli/bbl*linux* /usr/local/bin/bbl
mv bosh-cli/bosh*linux* /usr/local/bin/bosh

chmod +x /usr/local/bin/*

cd bbl-state

bbl print-env
#eval "$(bbl print-env)"

#bosh -d credhub deploy ${BASEDIR}/source/integration-tests/manifest/credhub.yml