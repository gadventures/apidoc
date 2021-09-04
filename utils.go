package apidoc

import "reflect"

// copy utils

func copySlice(s []interface{}) []interface{} {
	c := make([]interface{}, len(s))
	for i := 0; i < len(s); i++ {
		c[i] = copyValue(s[i])
	}
	return c
}

func copyValue(v interface{}) interface{} {
	switch t := v.(type) {
	case Document:
		return *t.Copy() // there should never be a nil Document
	case []interface{}:
		return copySlice(t)
	default:
		return t
	}
}

// equality utils

func docEqual(k string, d Document, o interface{}) bool {
	doc, ok := o.(Document)
	return ok && d.Equal(doc)
}

func sliceEqual(k string, s []interface{}, o interface{}) bool {
	slc, ok := o.([]interface{})
	if !ok {
		return false
	}
	if len(s) != len(slc) {
		return false
	}
	for i := range s {
		if valEqual(k, s[i], slc[i]) {
			continue
		}
		return false
	}
	return true
}

func valEqual(k string, v, o interface{}) bool {
	switch t := v.(type) {
	case bool, float64, string:
		return reflect.DeepEqual(t, o)
	case Document:
		return docEqual(k, t, o)
	case []interface{}:
		return sliceEqual(k, t, o)
	case nil:
		return o == nil
	default:
		return false
	}
}
