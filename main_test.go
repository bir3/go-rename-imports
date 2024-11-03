package main

//go:generate testdata/generate-tests.sh generated_test.go

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"sigs.k8s.io/yaml"
)

type Spec struct {
	Input  string `json:"input"` // Affects YAML field names too.
	Cmd    string `json:"cmd"`
	Expect string `json:"expect"`
}

func parse(t *testing.T, s string) Spec {
	var test Spec
	err := yaml.Unmarshal([]byte(s), &test)
	if err != nil {
		t.Fatal(err)
	}
	return test
	/*
		i := strings.Index(s, substr)
		if i < 0 {
			t.Fatalf("did not find %s", substr)
		}
		lines := s[i:]
		k := strings.Index(lines, "\n")
		if k < 0 {
			t.Fatalf("missing newline after %s", substr)
		}
		rest := lines[k+1:]
		k2 := strings.Index(rest, "\n---")
		if k2 < 0 {
			t.Fatal("missing final sentinel string '---'")
		}
		return lines[0:k], rest[0 : k2+1]
	*/
}

func writeFile(t *testing.T, filepath string, data string) {
	err := os.WriteFile(filepath, []byte(data), 0644)
	if err != nil {
		t.Fatal(err)
	}
}

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags

	buildCmd := exec.Command("go", "build")
	_, err := buildCmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}

	m.Run()
}

func runTest(t *testing.T, testfile string) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	debug := testing.Verbose()
	var dir string
	if debug {
		os.MkdirAll("tmp", 0755)
		dir = "tmp"
	} else {
		dir = t.TempDir()
	}

	data, err := os.ReadFile(testfile)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	test := parse(t, s)

	fields := strings.Fields(test.Cmd)
	assert(len(fields) > 2)

	toolPath := filepath.Join(wd, "go-rename-imports") // built by TestMain

	writeFile(t, filepath.Join(dir, "input.go"), test.Input)
	writeFile(t, filepath.Join(dir, "actual.go"), test.Input) // will modify
	writeFile(t, filepath.Join(dir, "expect.go"), test.Expect)

	gofile := filepath.Join(dir, "actual.go")

	if fields[0] != "go-rename-imports" {
		t.Fatal("test cmd must begin with go-rename-imports")
	}
	fields[0] = toolPath
	fields = append(fields, "actual.go")

	if testing.Verbose() {
		fmt.Println("****", testfile)
		fmt.Println("*** CMD:", fields)
	}
	cmd := exec.Command(fields[0], fields[1:]...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s : %s", out, err)
	}
	buf, err := os.ReadFile(gofile)
	if err != nil {
		t.Fatal(err)
	}
	actualOutput := strings.TrimSpace(string(buf))

	// HACK: output uses tab for indent but our expected string uses spaces
	actualOutput = strings.ReplaceAll(actualOutput, "\n\t", "\n  ")
	if actualOutput != strings.TrimSpace(test.Expect) {
		s := "** expected:\n" + test.Expect
		s += "** actual:\n" + actualOutput
		s += "\n"
		t.Fatalf(s)
	}
}
