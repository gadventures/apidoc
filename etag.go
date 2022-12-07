package apidoc

import (
	"fmt"
	"hash/fnv"
)

// ETag is type for storing the calculated etag/checksum of a Document
type ETag uint64

func (e ETag) String() string {
	return fmt.Sprintf("%x", uint64(e))
}

// NewETag returns new ETag initialized with provided hexadecimal string
func NewETag(s string) (ETag, error) {
	var e uint64
	if _, err := fmt.Sscanf(s, "%x", &e); err != nil {
		return ETag(e), err
	}
	return ETag(e), nil
}

// ETag returns the checksum of the document
func (d Document) ETag() (ETag, error) {
	h := fnv.New64a()
	err := encodeDocument(h, d, true)
	return ETag(h.Sum64()), err
}
