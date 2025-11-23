package storage

import "errors"

// ErrNotFound is returned by repositories when entity does not exist.
var ErrNotFound = errors.New("not found")
