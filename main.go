package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

// https://eli.thegreenplace.net/2021/rewriting-go-source-code-with-ast-tooling/

// https://pkg.go.dev/golang.org/x/tools@v0.17.0/go/ast/astutil

func exitOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s : %s", msg, err)
	}
}
func assert(p bool) {
	if !p {
		panic("assert failed")
	}
}

func getImports(goFile *ast.File) []string {
	out := []string{}
	for _, imp := range goFile.Imports {
		pkg := imp.Path.Value
		pkg = pkg[1 : len(pkg)-1]
		if pkg != "C" {
			out = append(out, pkg)
		}
	}
	return out
}

func listImports(path string, showPath bool) {
	fset := token.NewFileSet()
	goFile, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
	// parser.ImportsOnly : only for reading, can't emit file in that case
	exitOnError(err, fmt.Sprintf("file %s", path[0:len(path)-1]))

	for _, pkg := range getImports(goFile) {
		if showPath {
			fmt.Printf("%s %s\n", pkg, path)
		} else {
			fmt.Printf("%s\n", pkg)
		}
	}
}

func modifyImports(cmd string, path string, args []string, write bool) {
	fset := token.NewFileSet()
	goFile, err := parser.ParseFile(fset, path, nil, parser.ParseComments)

	if err != nil {
		log.Fatalf("failed to parse %s : %s", path, err)
	}

	imports := make(map[string]bool)
	for _, p := range getImports(goFile) {
		imports[p] = true
	}
	modified := false
	for _, s := range args {
		e := strings.Split(s, "|")
		pkg := e[0]
		// catch input errors
		assert(len(pkg) > 0)
		assert(pkg[0] != '"')
		assert(pkg[len(pkg)-1] != '"')
		assert(strings.TrimSpace(pkg) == pkg)
		assert(len(strings.Fields(pkg)) == 1)
		switch cmd {
		case "rename-imports":
			assert(len(e) == 2)
			newPkg := e[1]
			if astutil.RewriteImport(fset, goFile, pkg, newPkg) {
				modified = true
			}
		case "rename-prefix-imports":
			assert(len(e) == 2)
			// order is important:
			// 		rename internal/syscall/ => export/syscall/
			//      rename internal/         => special/
			// => we remove import once we have renamed it via dlist
			dlist := []string{}
			newPrefix := e[1]
			for ip, _ := range imports {
				if strings.HasPrefix(ip, pkg) {
					newPkg := newPrefix + ip[len(pkg):]
					if astutil.RewriteImport(fset, goFile, ip, newPkg) {
						modified = true
					} else {
						panic("must find")
					}
					dlist = append(dlist, ip)
				}
			}
			for _, ip := range dlist {
				delete(imports, ip)
			}
		case "delete-imports":
			assert(len(e) == 1)
			if astutil.DeleteImport(fset, goFile, pkg) {
				modified = true
			}
		case "add-imports":
			assert(len(e) == 1)
			astutil.AddImport(fset, goFile, pkg)
			modified = true
		default:
			log.Fatalf("unknown cmd %s", cmd)
		}
	}
	var buf bytes.Buffer
	printer.Fprint(&buf, fset, goFile)

	if write && modified {
		outfile := fmt.Sprintf("%s.tmp", path)
		err = os.WriteFile(outfile, buf.Bytes(), 0644)
		if err != nil {
			log.Fatalf("write %s : %s", outfile, err)
		}

		err = os.Rename(outfile, path)
		if err != nil {
			log.Fatalf("rename %s -> %s : %s", outfile, path, err)
		}

	}

}
func findGoFiles(dir string) []string {
	out := []string{}
	fileSystem := os.DirFS(dir)

	err := fs.WalkDir(fileSystem, ".", func(fspath string, entry fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		name := entry.Name()
		if entry.IsDir() && name == "testdata" {
			return fs.SkipDir
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		if strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go") {
			out = append(out, filepath.Join(dir, fspath))
		}
		return nil
	})
	if err != nil {
		log.Fatalf("failed with %s", err)
	}
	return out
}

func arg2files(args []string) []string {
	var out []string
	if len(args) == 0 {
		printUsage()
		log.Fatal("ERROR: missing <file/dir> ...")
	}
	for _, x := range args {
		info, err := os.Stat(x)
		if err != nil {
			log.Fatal(err)
		}
		if info.IsDir() {
			flist := findGoFiles(x)
			out = append(out, flist...)
		} else {
			// NOTE: we do not filter out non-go files here,
			// assume input is ok
			out = append(out, x)
		}
	}
	return out
}

type PatternList []string

func (i *PatternList) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *PatternList) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func printUsage() {
	s := `
modify go imports
usage:
  $self add-imports           [-w] -e pkg <file/dir>  ...
  $self delete-imports        [-w] -e pkg <file/dir>  ...
  $self rename-imports        [-w] -e pkg|newPkg <file/dir>  ...
  $self rename-prefix-imports [-w] -e pkgPrefix|newPrefix <file/dir>  ...
  $self find-go-files <file/dir> ..
  $self list-imports <file/dir> ..
-w = modify file in-place
-e <pattern> = can be given multiple times
<file/dir> = can be given multiple times
`
	s = strings.ReplaceAll(s, "$self", "rename")
	fmt.Fprint(os.Stderr, s)
}
func main() {
	if len(os.Args) == 1 {
		printUsage()
		os.Exit(1)
	}
	opt := func(s string) string {
		assert(s[0] == '-')
		assert(s[1] != '-')
		return s[1:]
	}

	fs := flag.NewFlagSet(os.Args[1], flag.ContinueOnError)
	cmd := os.Args[1]
	switch cmd {
	case "list-imports":
		var showPath bool
		fs.BoolVar(&showPath, opt("-show-path"), false, "include filepath")
		err := fs.Parse(os.Args[2:])
		if err != nil {
			printUsage()
			os.Exit(1)
		}
		for _, f := range arg2files(fs.Args()) {
			listImports(f, showPath)
		}

	case "find-go-files":
		err := fs.Parse(os.Args[2:])
		if err != nil {
			printUsage()
			os.Exit(1)
		}
		for _, f := range arg2files(fs.Args()) {
			fmt.Printf("%s\n", f)
		}

	case "add-imports",
		"delete-imports",
		"rename-imports",
		"rename-prefix-imports":
		var write bool

		var patterns PatternList
		fs.BoolVar(&write, opt("-w"), false, "write file in-place")
		fs.Var(&patterns, opt("-e"), "pattern, can be given multiple times")
		err := fs.Parse(os.Args[2:])
		if err != nil {
			printUsage()
			os.Exit(1)
		}
		if len(patterns) == 0 {
			printUsage()
			log.Fatal("ERROR: missing pattern -e <pattern>")
		}
		for _, path := range arg2files(fs.Args()) {
			modifyImports(cmd, path, patterns, write)
		}

	default:
		printUsage()
		log.Fatalf("ERROR: unknown command %s", cmd)
	}
}
