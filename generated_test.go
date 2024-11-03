package main
//
// DO NOT MODIFY - generated code via 'go generate'
//
import "testing"

func Test_add(t *testing.T) {
	runTest(t, "testdata/add.yaml")
}

func Test_delete(t *testing.T) {
	runTest(t, "testdata/delete.yaml")
}

func Test_rename_prefix(t *testing.T) {
	runTest(t, "testdata/rename-prefix.yaml")
}

func Test_rename(t *testing.T) {
	runTest(t, "testdata/rename.yaml")
}

