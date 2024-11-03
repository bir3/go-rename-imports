#! /usr/bin/env bash

set -eu


function emit() {
    cat <<END
package main

import "testing"

END
    for f in testdata/*.yaml
    do
        b=$(basename $f .yaml|tr '-' '_')
        cat <<END
func Test_$b(t *testing.T) {
	runTest(t, "$f")
}

END
    done
}

emit >$1
