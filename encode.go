package apidoc

import (
	"encoding/binary"
	"fmt"
	"io"
)

type encodeType uint8

const (
	encodeTypeInvalid encodeType = iota
	encodeTypeString
	encodeTypeBool
	encodeTypeFloat64
	encodeTypeListStart
	encodeTypeListEnd
	encodeTypeDocumentStart
	encodeTypeDocumentEnd
	encodeTypeNil
)

// ByteValue return the byte value of the encodeType
func (e encodeType) ByteValue() []byte {
	// we use [1]byte to ensure that the value is encoded onto a single byte
	byteVal := [1]byte{byte(e)}
	return byteVal[:]
}

// String representation of encodeType
func (e encodeType) String() string {
	switch e {
	case encodeTypeString:
		return "String"
	case encodeTypeBool:
		return "Bool"
	case encodeTypeFloat64:
		return "Float64"
	case encodeTypeListStart:
		return "ListStart"
	case encodeTypeListEnd:
		return "ListEnd"
	case encodeTypeDocumentStart:
		return "DocumentStart"
	case encodeTypeDocumentEnd:
		return "DocumentEnd"
	case encodeTypeNil:
		return "Nil"
	default:
		return fmt.Sprintf("Unknown(%d)", e)
	}
}

func encodeEncodeType(w io.Writer, e encodeType) error {
	if l, err := w.Write(e.ByteValue()); l != 1 {
		if err != nil {
			return err
		}
		return fmt.Errorf("encodeEDT fail for %s", e)
	}
	return nil
}

func encodeBool(w io.Writer, b bool) error {
	var v uint8
	if b {
		v = 1
	}
	_, err := w.Write([]byte{byte(encodeTypeBool), byte(v)})
	return err
}

func encodeDocument(w io.Writer, doc Document, sortKeys bool) error {
	if err := encodeEncodeType(w, encodeTypeDocumentStart); err != nil {
		return err
	}

	// extract the Document keys and sort them if necessary
	var keys []string
	if sortKeys {
		keys = doc.KeysSorted()
	} else {
		keys = doc.Keys()
	}

	for _, key := range keys {
		err := encodeString(w, key)
		if err != nil {
			return err
		}

		val := doc[key]
		if val == nil {
			err = encodeNil(w)
		} else {
			switch val := val.(type) {
			case string:
				err = encodeString(w, val)
			case float64:
				err = encodeFloat64(w, val)
			case bool:
				err = encodeBool(w, val)
			case Document:
				err = encodeDocument(w, val, sortKeys)
			case []interface{}:
				err = encodeList(w, val, sortKeys)
			default:
				return fmt.Errorf(
					"key %s has unexpected type %T for value %v",
					key, val, val,
				)
			}
		}
		if err != nil {
			return err
		}
	}
	return encodeEncodeType(w, encodeTypeDocumentEnd)
}

func encodeFloat64(w io.Writer, num float64) error {
	if err := encodeEncodeType(w, encodeTypeFloat64); err != nil {
		return err
	}
	return binary.Write(w, binary.LittleEndian, num)
}

func encodeList(w io.Writer, list []interface{}, sortKeys bool) error {
	if err := encodeEncodeType(w, encodeTypeListStart); err != nil {
		return err
	}

	var err error
	for idx, val := range list {
		// special case nil
		if val == nil {
			return encodeNil(w)
		}
		// handled types
		switch val := val.(type) {
		case bool:
			err = encodeBool(w, val)
		case float64:
			err = encodeFloat64(w, val)
		case string:
			err = encodeString(w, val)
		case Document:
			err = encodeDocument(w, val, sortKeys)
		case []interface{}:
			err = encodeList(w, val, sortKeys)
		default:
			return fmt.Errorf(
				"item at index %d has unexpected type %T for value %v",
				idx, val, val,
			)
		}
		if err != nil {
			return err
		}
	}
	return encodeEncodeType(w, encodeTypeListEnd)
}

func encodeNil(w io.Writer) error {
	return encodeEncodeType(w, encodeTypeNil)
}

func encodeString(w io.Writer, s string) error {
	if err := encodeEncodeType(w, encodeTypeString); err != nil {
		return err
	}

	data := []byte(s)
	if err := binary.Write(w, binary.LittleEndian, int64(len(data))); err != nil {
		return err
	}
	_, err := w.Write(data)
	return err
}

// decode here

func decodeEncodeType(r io.Reader) (encodeType, error) {
	buf := make([]byte, 1)
	l, err := r.Read(buf)
	if l != len(buf) {
		if err != nil {
			return encodeTypeInvalid, err
		}
		// NOTE: len of bytes just recursively calls ourself
		//       should be safe tail recursion
		return decodeEncodeType(r)
	}
	return encodeType(buf[0]), nil
}

func nextItem(r io.Reader) (encodeType, interface{}, error) {
	var val interface{}

	typ, err := decodeEncodeType(r)
	if err != nil {
		return typ, nil, err
	}

	switch typ {
	case encodeTypeDocumentStart:
	case encodeTypeDocumentEnd:
	case encodeTypeString:
		val, err = decodeString(r)
	case encodeTypeBool:
		val, err = decodeBool(r)
	case encodeTypeFloat64:
		val, err = decodeFloat64(r)
	case encodeTypeListStart:
	case encodeTypeListEnd:
	case encodeTypeNil:
		return typ, nil, nil
	default:
		err = fmt.Errorf("decoding not supported for type %v", typ)
	}
	return typ, val, err
}

func decodeDocument(r io.Reader) (Document, error) {
	doc := make(Document)
Loop:
	for {
		typ, val, err := nextItem(r)
		if err != nil {
			return doc, err
		}
		switch typ {
		case encodeTypeDocumentEnd:
			break Loop
		case encodeTypeString:
			key := val.(string)
			value, err := decodeValue(r)
			if err != nil {
				return doc, err
			}
			doc[key] = value
		default:
			return doc, fmt.Errorf(
				"top level decoding not supported for type %v",
				typ)
		}
	}
	return doc, nil
}

func decodeValue(r io.Reader) (interface{}, error) {
	typ, val, err := nextItem(r)
	if err != nil {
		return nil, err
	}
	switch typ {
	case encodeTypeDocumentStart:
		val, err = decodeDocument(r)
	case encodeTypeListStart:
		val, err = decodeList(r)
	case encodeTypeString:
	case encodeTypeBool:
	case encodeTypeFloat64:
	case encodeTypeNil:
		return nil, nil
	default:
		return nil, fmt.Errorf(
			"top level decoding not supported for type %v",
			typ)
	}
	return val, err
}

func decodeString(r io.Reader) (string, error) {
	var length, bytesRead int64
	err := binary.Read(r, binary.LittleEndian, &length)
	if err != nil {
		return "", err
	}
	raw := make([]byte, length)
	for bytesRead < length {
		data := raw[bytesRead:]
		l, err := r.Read(data)
		bytesRead += int64(l)
		if l < 1 && err != nil {
			return "", err
		}
	}
	return string(raw), nil
}

func decodeBool(r io.Reader) (bool, error) {
	raw := make([]byte, 1)
	l, err := r.Read(raw)
	if len(raw) != l {
		if err != nil {
			return false, err
		}
		// efficient tail recursive call in case data was not ready
		return decodeBool(r)
	}
	if uint8(raw[0]) == 1 {
		return true, nil
	}
	return false, nil
}

func decodeFloat64(r io.Reader) (float64, error) {
	var num float64
	err := binary.Read(r, binary.LittleEndian, &num)
	if err != nil {
		return 0, err
	}
	return num, nil
}

func decodeList(r io.Reader) ([]interface{}, error) {
	var list []interface{}

	for {
		encType, item, err := nextItem(r)
		if err != nil {
			return nil, err
		}
		// check the encodeType and act accordingly
		switch encType {
		case encodeTypeListEnd:
			return list, nil
		case encodeTypeDocumentStart:
			item, err = decodeDocument(r)
		case encodeTypeListStart:
			item, err = decodeList(r)
		case encodeTypeString:
		case encodeTypeBool:
		case encodeTypeFloat64:
		case encodeTypeNil:
		default:
			return nil, fmt.Errorf("top level decoding not supported for type %v", encType)
		}
		// return early if we encountered an error while recursively decoding
		if err != nil {
			return nil, err
		}
		// append the item and continue
		list = append(list, item)
	}
}
