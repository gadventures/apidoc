package apidoc

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"reflect"
	"sort"

	"github.com/golang/snappy"
)

/*
Document represents the downloaded REST API document
Internally its contents would be
bool, for booleans
float64, for all numbers
string, for strings
[]interface{} for slices
Document for nested objects
nil for nils
*/
type Document map[string]interface{}

// New returns new Document
func New() Document {
	return make(Document)
}

func (d Document) String() string {
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

// Equal compares if two documents are the same
func (d Document) Equal(other Document) bool {
	documentEqual := func(k string, vv Document, ov interface{}) bool {
		if odoc, tpok := ov.(Document); tpok {
			return vv.Equal(odoc)
		}
		return false
	}
	var sliceEquals func(string, []interface{}, interface{}) bool
	valEval := func(k string, v, ov interface{}) bool {
		switch vv := v.(type) {
		case bool:
			if !reflect.DeepEqual(vv, ov) {
				return false
			}
		case float64:
			if !reflect.DeepEqual(vv, ov) {
				return false
			}
		case string:
			if !reflect.DeepEqual(vv, ov) {
				return false
			}
		case Document:
			if !documentEqual(k, vv, ov) {
				return false
			}
		case nil:
			if ov != nil {
				return false
			}
		case []interface{}:
			if !sliceEquals(k, vv, ov) {
				return false
			}
		default:
			return false
		}
		return true
	}
	sliceEquals = func(k string, mslc []interface{}, ov interface{}) bool {
		oslc, tpok := ov.([]interface{})
		if !tpok {
			return false
		}
		if len(mslc) != len(oslc) {
			return false
		}
		for i := range mslc {
			if !valEval(k, mslc[i], oslc[i]) {
				return false
			}
		}
		return true
	}
	// basic len check
	if len(d) != len(other) {
		return false
	}
	// main comparison loop sort keys first
	var sortedKeys []string
	for k := range d {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	for _, k := range sortedKeys {
		v := d[k]
		ov, prs := other[k]
		if !prs {
			return false
		}
		if !valEval(k, v, ov) {
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
	var copySlice func([]interface{}) []interface{}
	copyValue := func(v interface{}) interface{} {
		switch vv := v.(type) {
		case Document:
			return *vv.Copy() // there should never be nil so this should be ok
		case []interface{}:
			return copySlice(vv)
		default:
			return vv
		}
	}
	copySlice = func(slc []interface{}) []interface{} {
		newSlc := make([]interface{}, len(slc))
		for i := 0; i < len(slc); i++ {
			newSlc[i] = copyValue(slc[i])
		}
		return newSlc
	}
	doc := make(Document)
	for k, v := range *d {
		doc[k] = copyValue(v)
	}
	return &doc
}

// TraverseFunc is an alias for function that will be executed
// for every node in Document tree
type TraverseFunc func(*Document, string, interface{})

// TraverseCall will visit every node in Document tree
// and call provided f on each node
func (d Document) TraverseCall(f TraverseFunc) {
	// ok main loop
	for k, v := range d {
		switch vv := v.(type) {
		case Document:
			vv.TraverseCall(f)
		case []interface{}:
			// traverse documents if they were inside the slice
			for _, vvv := range vv {
				if vvvv, isDoc := vvv.(Document); isDoc {
					vvvv.TraverseCall(f)
				}
			}
			// call this anyway for mixed slices or slices of non Document
			f(&d, k, vv)
		default:
			f(&d, k, vv)
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
		return fmt.Errorf(
			"binary.Read of checksum failed: unmarshalling: %v", err)
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
			return fmt.Errorf("Expected Document got %T", val)
		}
		*d = doc
		return nil
	default:
		// must be legacy json then
	}
	if crc32.ChecksumIEEE(remdata) != oldcrc {
		return errors.New("checksum does not match - unmarshalling")
	}
	return json.Unmarshal(remdata, d)
}

// MarshalBinary allows documents to be stored in cache
func (d Document) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 512*len(d)))
	if err := binary.Write(
		buf, binary.LittleEndian, uint32(serBinary)); err != nil {
		return nil, errors.New(
			"failed to encode serializationType - marshalling")
	}
	w := snappy.NewBufferedWriter(buf)
	err := encodeDocument(w, d, false)
	if err != nil {
		return nil, err
	}
	err = w.Close()
	return buf.Bytes(), err
}

// ErrGAPI type represents error message as returned by GAPI
type ErrGAPI struct {
	URI            string
	HTTPStatusCode int
	Msg            string
}

func (r *ErrGAPI) Error() string {
	return fmt.Sprintf("GAPI error for %s HTTP Status %d %s", r.URI, r.HTTPStatusCode, r.Msg)
}

//GAPIError returns ErrGAPI if the document is an GAPI error
///nil otherwise
func (d Document) GAPIError(uri string) *ErrGAPI {
	statusCode, hasCode := d["http_status_code"]
	status, hasStatus := d["status"]
	message, hasMsg := d["message"]
	if hasCode && hasMsg {
		statusCode, hasCode := statusCode.(float64)
		message, hasMsg := message.(string)
		if hasCode && hasMsg {
			return &ErrGAPI{uri, int(statusCode), message}
		}
	} else if hasStatus && hasMsg {
		statusCode, hasCode := status.(float64)
		message, hasMsg := message.(string)
		if hasCode && hasMsg {
			return &ErrGAPI{uri, int(statusCode), message}
		}
	}
	return nil
}

// GetPath recursively searches for value at provided path
// e.g. GetPath("staff_profiles", "id") would return attribute id in
// /staff_profiles/id
func (d Document) GetPath(parts ...string) (interface{}, bool) {
	switch len(parts) {
	case 0:
		return "", false
	case 1:
		v, prs := d[parts[0]]
		return v, prs
	}
	// ok there is more parts
	// for now only follow path of docs of docs (no lists)
	nextDoc, castOK := d[parts[0]].(Document)
	if !castOK {
		return "", false
	}
	return nextDoc.GetPath(parts[1:]...)
}
