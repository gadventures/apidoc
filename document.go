package apidoc

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"sort"

	"github.com/golang/snappy"
)

// Document represents a single G API resource. Internally its contents
// would be represented by Go primitives.
//
// bool          for booleans
// float64       for numbers
// string        for strings
// []interface{} for arrays
// Document      for nested objects
// nil           for nulls
type Document map[string]interface{}

// New returns new Document
func New() Document {
	return make(Document)
}

// Keys is a convenience method to return the keys in this Document
func (d Document) Keys() []string {
	keys := make([]string, 0, len(d))
	for k := range d {
		keys = append(keys, k)
	}
	return keys
}

// KeysSorted is a convenience method to return the keys in this Document, in sorted
// order
func (d Document) KeysSorted() []string {
	keys := d.Keys()
	sort.Strings(keys)
	return keys
}

// String satisifies the Stringer interface
func (d Document) String() string {
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(data)
}

// Equal compares if two Documents are the same
func (d Document) Equal(other Document) bool {
	// basic len check
	if len(d) != len(other) {
		return false
	}
	// main comparison loop sort keys first
	for _, k := range d.KeysSorted() {
		v := d[k]
		o, ok := other[k]
		if !ok {
			return false
		}
		if !valueEquals(k, v, o) {
			return false
		}
	}
	return true
}

// Copy creates a new copy of document
func (d *Document) Copy() *Document {
	if d == nil {
		return nil
	}
	doc := make(Document)
	for k, v := range *d {
		doc[k] = copyValue(v)
	}
	return &doc
}

// TraverseFunc is an alias for a function that will be executed for every item
// in the Document
type TraverseFunc = func(*Document, string, interface{})

// TraverseCall will visit every item in Document and call the provided
// TraverseFunc on each item
func (d Document) TraverseCall(f TraverseFunc) {
	for k, v := range d {
		switch t := v.(type) {
		case Document:
			t.TraverseCall(f)
		case []interface{}:
			// traverse documents if they were inside the slice
			for _, item := range t {
				if doc, isDoc := item.(Document); isDoc {
					doc.TraverseCall(f)
				}
			}
			// call this anyway for mixed slices or slices of non Document
			f(&d, k, t)
		default:
			f(&d, k, t)
		}
	}
}

type serializationType uint32

const (
	serInvalid serializationType = iota
	serBinary                    // this is the snappy encoded version
)

// UnmarshalBinary implements binary decoding
func (d *Document) UnmarshalBinary(data []byte) error {
	var oldcrc uint32
	buf := bytes.NewBuffer(data)
	err := binary.Read(buf, binary.LittleEndian, &oldcrc)
	if err != nil {
		return fmt.Errorf("binary.Read of checksum failed: unmarshaling: %w", err)
	}
	remdata := buf.Bytes()
	switch serializationType(oldcrc) {
	case serBinary:
		val, err := decodeValue(snappy.NewReader(bytes.NewReader(remdata)))
		if err != nil {
			return err
		}
		doc, ok := val.(Document)
		if !ok {
			return fmt.Errorf("expected Document got %T", val)
		}
		*d = doc
		return nil
	default:
		// must be legacy json then
	}
	if crc32.ChecksumIEEE(remdata) != oldcrc {
		return errors.New("checksum does not match - unmarshaling")
	}
	return json.Unmarshal(remdata, d)
}

// MarshalBinary allows documents to be stored in cache
func (d Document) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 512*len(d)))
	if err := binary.Write(
		buf, binary.LittleEndian, uint32(serBinary)); err != nil {
		return nil, errors.New("failed to encode serializationType - marshaling")
	}
	w := snappy.NewBufferedWriter(buf)
	err := encodeDocument(w, d, false)
	if err != nil {
		return nil, err
	}
	err = w.Close()
	return buf.Bytes(), err
}

// ErrGAPI type represents error response returned by GAPI
//
// See: https://developers.gadventures.com/docs/rest.html#errors
type ErrGAPI struct {
	HTTPStatusCode int
	Message        string
	ErrorID        string
	URI            string
}

// Error satisfies the error interface for the ErrGAPI type
func (r *ErrGAPI) Error() string {
	return fmt.Sprintf("GAPI error for %s HTTP Status %d %s %s", r.URI, r.HTTPStatusCode, r.ErrorID, r.Message)
}

// GAPIError returns *ErrGAPI if the document is a GAPI error, nil otherwise
func (d Document) GAPIError(uri string) *ErrGAPI {
	if _, hasErr := d["error_id"]; hasErr {
		errID, _ := d["error_id"].(string)
		msg, _ := d["message"].(string)
		status, _ := d["http_status_code"].(int)
		return &ErrGAPI{
			ErrorID:        errID,
			HTTPStatusCode: status,
			Message:        msg,
			URI:            uri,
		}
	}
	return nil
}

// GetPath recursively searches for value at provided path
//
// e.g. GetPath("staff_profiles", "id") would return attribute id in
// /staff_profiles/id
func (d Document) GetPath(parts ...string) (interface{}, bool) {
	switch len(parts) {
	case 0:
		return "", false
	case 1:
		v, prs := d[parts[0]]
		return v, prs
	default:
		// following a nested path
		//
		// NOTE: we are not dealing with lists at this time
		doc, ok := d[parts[0]].(Document)
		if !ok {
			return nil, false
		}
		return doc.GetPath(parts[1:]...)
	}
}
