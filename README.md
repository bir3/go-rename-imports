# go-rename-imports

Install: `go install github.com/bir3/go-rename-imports@latest`

```
go-rename-imports

modify go imports
usage:
  rename add-imports           [-w] -e pkg <file/dir>  ...
  rename delete-imports        [-w] -e pkg <file/dir>  ...
  rename rename-imports        [-w] -e pkg|newPkg <file/dir>  ...
  rename rename-prefix-imports [-w] -e pkgPrefix|newPrefix <file/dir>  ...
  rename find-go-files <file/dir> ..
  rename list-imports [-show-path] <file/dir> ..
-w = modify file in-place
-e <pattern> = can be given multiple times
<file/dir> = can be given multiple times```
