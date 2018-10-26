#!/bin/bash 

set -eu

BASEDIR=$(pwd)

mv bbl-cli/bbl*linux* /usr/local/bin/bbl
mv bosh-cli/bosh*linux* /usr/local/bin/bosh

chmod +x /usr/local/bin/*

cd bbl-state

eval "$(bbl print-env)"

for release in $(find ${BASEDIR} -name '*-bosh-release' -type d); do
    bosh upload-release --sha1="$(cat ${release}/sha1)" --version="$(cat ${release}/version)" "$(cat ${release}/url)"
done

for stemcell in $(find ${BASEDIR} -name '*-stemcell' -type d); do
    bosh upload-stemcell --sha1="$(cat ${stemcell}/sha1)" --version="$(cat ${stemcell}/version)" "$(cat ${stemcell}/url)"
done

#bosh -d credhub deploy ${BASEDIR}/source/integration-tests/manifest/credhub.yml