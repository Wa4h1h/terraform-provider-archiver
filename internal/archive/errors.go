package archive

import "errors"

var (
	ErrPathIsNotDir  = errors.New("path is not a directory path")
	ErrPathIsNotFile = errors.New("path is not a file path")
)
