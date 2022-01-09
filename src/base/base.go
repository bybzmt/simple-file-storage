package base

import (
	"path"
)

var BaseDir string = "./"
var Debug bool = false

func SafeFile(file string) string {
	f := path.Clean(file)

	return f
}
