# go-rename-imports

Install:
```
go install github.com/bir3/go-rename-imports@latest
```

```
go-rename-imports

modify go imports
	https://github.com/bir3/go-rename-imports
usage:
  go-rename-imports rename [-w] -e pkg|newPkg  <file/dir>  ...
  go-rename-imports rename [-w] -p pkgPrefix|newPrefix <file/dir>  ...
  go-rename-imports add    [-w] -e pkg <file/dir>  ...
  go-rename-imports delete [-w] -e pkg <file/dir>  ...
  go-rename-imports find-go-files <file/dir> ..
  go-rename-imports list-imports [-show-path] <file/dir> ..
-w = modify file in-place
-e pat / -p pat = can be given multiple times
-e pat / -p pat = can be mixed
<file/dir> = can be given multiple times
```

# Limitations

Only the imports are updated - not references to them.

# Links

- https://eli.thegreenplace.net/2021/rewriting-go-source-code-with-ast-tooling/
- https://pkg.go.dev/golang.org/x/tools@v0.17.0/go/ast/astutil
