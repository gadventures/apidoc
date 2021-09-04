package apidoc

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
)

// UnmarshalJSON implements json unmarshaling of Document
func (d *Document) UnmarshalJSON(data []byte) error {
	// we need to be in charge of unmarshaling to retain some sanity
	// everything we fetch will be json object -- i guess if wrong error will be thrown right away
	var root map[string]interface{}
	if err := json.Unmarshal(data, &root); err != nil {
		return err
	}
	/*
		To unmarshal JSON into an interface value, Unmarshal stores one of these in the interface value:
		bool, for JSON booleans
		float64, for JSON numbers
		string, for JSON strings
		[]interface{}, for JSON arrays
		map[string]interface{}, for JSON objects
		nil for JSON null
	*/
	var unpackObject func(map[string]interface{}) (Document, error)
	var unpackValue func(v interface{}) (interface{}, error)
	unpackList := func(l []interface{}) ([]interface{}, error) {
		list := make([]interface{}, len(l))
		for i, v := range l {
			uv, err := unpackValue(v)
			if err != nil {
				return list, err
			}
			list[i] = uv
		}
		return list, nil
	}
	unpackValue = func(v interface{}) (interface{}, error) {
		switch vv := v.(type) {
		case bool:
			return vv, nil
		case float64:
			return vv, nil
		case string:
			return vv, nil
		case []interface{}:
			u, err := unpackList(vv)
			if err != nil {
				return nil, err
			}
			return u, nil
		case map[string]interface{}:
			u, err := unpackObject(vv)
			if err != nil {
				return nil, err
			}
			return u, nil
		case nil:
			return nil, nil
		default:
			return nil, fmt.Errorf("illegal type %T", v)
		}
	}
	unpackObject = func(rawObj map[string]interface{}) (Document, error) {
		d := make(Document)
		for k, v := range rawObj {
			uv, err := unpackValue(v)
			if err != nil {
				return d, err
			}
			d[k] = uv
		}
		return d, nil
	}
	// ok do the root
	rd, rerr := unpackObject(root)
	if rerr == nil {
		*d = rd
	}
	return rerr
}

// MarshalJSON implements json marshaling of Document
func (d Document) MarshalJSON() ([]byte, error) {
	// we need to be in charge of marshaling as well because
	// go marshalls nil slices to null
	// where we want []
	var buf bytes.Buffer
	err := d.WriteOutJSON(&buf)
	return buf.Bytes(), err
}

// WriteOutJSON is more memory efficient alternative to MarshalJSON
// it so because it directly writes to io.Writer without a need
// to alocate another buffer
func (d Document) WriteOutJSON(w io.Writer) error {
	return jsonMarshalDocument(bufio.NewWriter(w), d, true)
}

func jsonMarshalDocument(w *bufio.Writer, doc Document, flush bool) error {
	// Supported values
	//
	// bool                    for JSON booleans
	// float64                 for JSON numbers
	// string                  for JSON strings
	// []interface{}           for JSON arrays
	// map[string]interface{}  for JSON objects
	// nil                     for JSON null

	// optionally flush the writer
	if flush {
		defer w.Flush()
	}

	// start serializing the document
	_, err := w.WriteRune('{')
	if err != nil {
		return err
	}

	for idx, key := range doc.KeysSorted() {
		// next JSON key/value
		if idx > 0 {
			_, err = w.WriteRune(',')
			if err != nil {
				return err
			}
		}
		// write the key
		err = jsonMarshalString(w, key)
		if err != nil {
			return err
		}
		w.WriteRune(':')
		// write the value
		val := doc[key]
		switch val := val.(type) {
		case string:
			err = jsonMarshalString(w, val)
		case float64:
			err = jsonMarshalFloat64(w, val)
		case bool:
			err = jsonMarshalBool(w, val)
		case Document:
			err = jsonMarshalDocument(w, val, false)
		case []interface{}:
			err = jsonMarshalList(w, val)
		case nil:
			err = jsonMarshalNil(w)
		default:
			return fmt.Errorf(
				"key %s has unexpected type %T for value %v",
				key, val, val)
		}
		if err != nil {
			return err
		}
	}
	_, err = w.WriteRune('}')
	return err
}

const (
	jsonQuote = '"'
	backSlash = '\\'
)

func jsonMarshalString(w *bufio.Writer, s string) error {
	_, err := w.WriteRune(jsonQuote)
	if err != nil {
		return err
	}
	for _, runeValue := range s {
		switch runeValue {
		case '\b':
			_, err = w.WriteString(`\b`)
		case '\n':
			_, err = w.WriteString(`\n`)
		case '\f':
			_, err = w.WriteString(`\f`)
		case '\r':
			_, err = w.WriteString(`\r`)
		case '\t':
			_, err = w.WriteString(`\t`)
		case jsonQuote:
			_, err = w.WriteString(`\"`)
		case backSlash:
			_, err = w.WriteString(`\\`)
		default:
			// if rune value is less than `space` 0x20 then per ASCII/UTF8 table
			// it is a control character that has no business being in json
			// so discard it
			if runeValue >= ' ' {
				_, err = w.WriteRune(runeValue)
			}
		}
		if err != nil {
			return err
		}
	}
	_, err = w.WriteRune(jsonQuote)
	return err
}

func jsonMarshalFloat64(w *bufio.Writer, n float64) error {
	if math.Ceil(n) == n {
		// integer
		fmt.Fprintf(w, "%.f", n)
	} else {
		// float
		fmt.Fprintf(w, "%f", n)
	}
	return nil
}

func jsonMarshalBool(w *bufio.Writer, b bool) error {
	var err error
	switch b {
	case true:
		_, err = w.WriteString("true")
	case false:
		_, err = w.WriteString("false")
	}
	return err
}

func jsonMarshalList(w *bufio.Writer, list []interface{}) error {
	var err error
	w.WriteRune('[')
	for idx, val := range list {
		if idx > 0 {
			w.WriteRune(',')
		}
		switch val := val.(type) {
		case string:
			err = jsonMarshalString(w, val)
		case float64:
			err = jsonMarshalFloat64(w, val)
		case bool:
			err = jsonMarshalBool(w, val)
		case Document:
			err = jsonMarshalDocument(w, val, false)
		case []interface{}:
			err = jsonMarshalList(w, val)
		case nil:
			err = jsonMarshalNil(w)
		default:
			return fmt.Errorf(
				"item at index %d has unexpected type %T for value %v",
				idx, val, val)
		}

		if err != nil {
			return err
		}
	}
	w.WriteRune(']')
	return nil
}

func jsonMarshalNil(w *bufio.Writer) error {
	w.WriteString("null")
	return nil
}
