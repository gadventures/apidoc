package apidoc

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sort"
)

type serDataType uint8

const (
	serDInvalid serDataType = iota
	serDString
	serDBool
	serDFloat64
	serDListStart
	serDListEnd
	serDDocumentStart
	serDDocumentEnd
	serDNil
)

func encodeSerType(w io.Writer, typ serDataType) error {
	var ary [1]byte
	ary[0] = byte(typ)
	if l, err := w.Write(ary[:]); l != 1 {
		if err != nil {
			return err
		}
		return errors.New("encodeSerType fail")
	}
	return nil
}

func encodeDocument(w io.Writer, doc Document, sortKeys bool) error {
	if err := encodeSerType(w, serDDocumentStart); err != nil {
		return err
	}
	keys := make([]string, 0, len(doc))
	for k := range doc {
		keys = append(keys, k)
	}
	if sortKeys {
		sort.Strings(keys)
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
					"Key %s has unexpected type %T for value %v",
					key, val, val)
			}
		}
		if err != nil {
			return err
		}
	}
	if err := encodeSerType(w, serDDocumentEnd); err != nil {
		return err
	}
	return nil
}

func encodeString(w io.Writer, str string) error {
	err := encodeSerType(w, serDString)
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

func encodeNil(w io.Writer) error {
	return encodeSerType(w, serDNil)
}

func encodeFloat64(w io.Writer, num float64) error {
	err := encodeSerType(w, serDFloat64)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, num)
	return err
}

func encodeBool(w io.Writer, b bool) error {
	var v uint8
	if b {
		v = 1
	}
	_, err := w.Write([]byte{byte(serDBool), byte(v)})
	return err
}

func encodeList(w io.Writer, list []interface{}, sortKeys bool) error {
	err := encodeSerType(w, serDListStart)
	if err != nil {
		return err
	}

	for idx, val := range list {
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
					"Item at index %d has unexpected type %T for value %v",
					idx, val, val)
			}
		}
		if err != nil {
			return err
		}
	}
	err = encodeSerType(w, serDListEnd)
	return err
}

// decode here

func decodeSerType(r io.Reader) (serDataType, error) {
	buf := make([]byte, 1)
	l, err := r.Read(buf)
	if l != len(buf) {
		if err != nil {
			return serDInvalid, err
		}
		// not len of bytes just recursively call outselfs
		// should be safe tail recursion
		return decodeSerType(r)
	}
	return serDataType(buf[0]), nil
}

func nextItem(r io.Reader) (serDataType, interface{}, error) {
	var val interface{}
	typ, err := decodeSerType(r)
	if err != nil {
		return typ, nil, err
	}
	switch typ {
	case serDDocumentStart:
	case serDDocumentEnd:
	case serDString:
		val, err = decodeString(r)
	case serDBool:
		val, err = decodeBool(r)
	case serDFloat64:
		val, err = decodeFloat64(r)
	case serDListStart:
	case serDListEnd:
	case serDNil:
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
		case serDDocumentEnd:
			break Loop
		case serDString:
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
	case serDDocumentStart:
		val, err = decodeDocument(r)
	case serDListStart:
		val, err = decodeList(r)
	case serDString:
	case serDBool:
	case serDFloat64:
	case serDNil:
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
		case serDListEnd:
			break Loop
		case serDDocumentStart:
			val, err = decodeDocument(r)
		case serDListStart:
			val, err = decodeList(r)
		case serDString:
		case serDBool:
		case serDFloat64:
		case serDNil:
		default:
			return lst, fmt.Errorf(
				"top level decoding not supported for type %v",
				typ)
		}
		lst = append(lst, val)
	}
	return lst, nil
}
