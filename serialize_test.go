package apidoc

import (
	"encoding/json"
	"testing"
)

func TestEmptyList(t *testing.T) {
	doc := New()
	var vodka []interface{}
	doc["party"] = vodka

	if l, ok := doc["party"].([]interface{}); !ok || len(l) != 0 {
		t.Errorf("Expected empty list")
	}

	// more types for fun
	doc["intnum"] = 6.0
	doc["floatnum"] = 6.5
	doc["txt"] = `bam"bam`

	data, err := json.Marshal(&doc)
	if err != nil {
		t.Error(err)
	}

	// test json encoding
	obtained := string(data)
	expected := `{"floatnum":6.500000,"intnum":6,"party":[],"txt":"bam\"bam"}`
	if obtained != expected {
		t.Errorf("Expected %s but got %s", expected, obtained)
	}
}
