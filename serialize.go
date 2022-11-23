package apidoc

import (
	"encoding/binary"
	"fmt"
	"io"
)

type serializeDataType uint8

const (
	sdtInvalid serializeDataType = iota
	sdtString
	sdtBool
	sdtFloat64
	sdtListStart
	sdtListEnd
	sdtDocumentStart
	sdtDocumentEnd
	sdtNil
)

func (s serializeDataType) String() string {
	switch s {
	case sdtString:
		return "String"
	case sdtBool:
		return "Bool"
	case sdtFloat64:
		return "Float64"
	case sdtListStart:
		return "ListStart"
	case sdtListEnd:
		return "ListEnd"
	case sdtDocumentStart:
		return "DocumentStart"
	case sdtDocumentEnd:
		return "DocumentEnd"
	case sdtNil:
		return "Nil"
	default:
		panic(fmt.Sprintf("unrecognized serializeDataType %d %T %#v", s, s, s))
	}
}

func encodeSDT(w io.Writer, t serializeDataType) error {
	var ary [1]byte = [1]byte{byte(t)}
	if l, err := w.Write(ary[:]); l != 1 {
		if err != nil {
			return err
		}
		return fmt.Errorf("encodeSDT fail for %s", t)
	}
	return nil
}

func encodeBool(w io.Writer, b bool) error {
	var v uint8
	if b {
		v = 1
	}
	_, err := w.Write([]byte{byte(sdtBool), byte(v)})
	return err
}

func encodeDocument(w io.Writer, doc Document, sortKeys bool) error {
	err := encodeSDT(w, sdtDocumentStart)
	if err != nil {
		return err
	}

	// extract the Document keys
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
	return encodeSDT(w, sdtDocumentEnd)
}

func encodeFloat64(w io.Writer, num float64) error {
	err := encodeSDT(w, sdtFloat64)
	if err != nil {
		return err
	}
	return binary.Write(w, binary.LittleEndian, num)
}

func encodeList(w io.Writer, list []interface{}, sortKeys bool) error {
	err := encodeSDT(w, sdtListStart)
	if err != nil {
		return err
	}

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
	return encodeSDT(w, sdtListEnd)
}

func encodeNil(w io.Writer) error {
	return encodeSDT(w, sdtNil)
}

func encodeString(w io.Writer, str string) error {
	err := encodeSDT(w, sdtString)
	if err != nil {
		return err
	}
	data := []byte(str)
	err = binary.Write(w, binary.LittleEndian, int64(len(data)))
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

// decode here

func decodeSDT(r io.Reader) (serializeDataType, error) {
	buf := make([]byte, 1)
	l, err := r.Read(buf)
	if l != len(buf) {
		if err != nil {
			return sdtInvalid, err
		}
		// not len of bytes just recursively call outselfs
		// should be safe tail recursion
		return decodeSDT(r)
	}
	return serializeDataType(buf[0]), nil
}

func nextItem(r io.Reader) (serializeDataType, interface{}, error) {
	var val interface{}
	typ, err := decodeSDT(r)
	if err != nil {
		return typ, nil, err
	}
	switch typ {
	case sdtDocumentStart:
	case sdtDocumentEnd:
	case sdtString:
		val, err = decodeString(r)
	case sdtBool:
		val, err = decodeBool(r)
	case sdtFloat64:
		val, err = decodeFloat64(r)
	case sdtListStart:
	case sdtListEnd:
	case sdtNil:
		return typ, nil, nil
	default:
		err = fmt.Errorf(
			"decoding not supported for type %v",
			typ)
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
		case sdtDocumentEnd:
			break Loop
		case sdtString:
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
	case sdtDocumentStart:
		val, err = decodeDocument(r)
	case sdtListStart:
		val, err = decodeList(r)
	case sdtString:
	case sdtBool:
	case sdtFloat64:
	case sdtNil:
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
	var lst []interface{}
Loop:
	for {
		typ, val, err := nextItem(r)
		if err != nil {
			return lst, err
		}
		switch typ {
		case sdtListEnd:
			break Loop
		case sdtDocumentStart:
			val, _ = decodeDocument(r)
		case sdtListStart:
			val, _ = decodeList(r)
		case sdtString:
		case sdtBool:
		case sdtFloat64:
		case sdtNil:
		default:
			return lst, fmt.Errorf(
				"top level decoding not supported for type %v",
				typ)
		}
		lst = append(lst, val)
	}
	return lst, nil
}
