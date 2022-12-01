package apidoc

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
