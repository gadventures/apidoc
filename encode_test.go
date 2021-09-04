package apidoc

import (
	"bytes"
	"encoding/json"
	"testing"
)

var sampleLargeDoc Document

func init() {
	sampleLargeDoc = bigSampleDoc(500)
}

func sampleDoc() Document {
	doc := make(Document)
	doc["name"] = "Jon Vain"
	doc["income"] = 14900.58
	doc["employed"] = true
	doc["notoriety"] = 5.0
	doc["useful"] = nil
	doc["friends"] = []interface{}{
		"Joe Strong", "Tim Bleak", "Judith Believable",
	}
	doc["matrix"] = []interface{}{
		[]interface{}{5.0, 4.9, 8.7},
		[]interface{}{9.0, 11.3, 15.88, 7.11},
	}
	return doc
}

func bigSampleDoc(depth int) Document {
	size := 500
	doc := make(Document)
	chlds := make([]interface{}, size)
	for i := 0; i < size; i++ {
		chlds[i] = sampleDoc()
	}
	doc["children"] = chlds
	if depth > 0 {
		doc["foo"] = bigSampleDoc(depth - 1)
	}
	return doc
}

func TestEncodeDecode(t *testing.T) {
	for _, doc := range []Document{sampleDoc(), bigSampleDoc(3)} {
		network := new(bytes.Buffer)
		err := encodeDocument(network, doc, false)
		if err != nil {
			t.Error(err)
		}
		v, err := decodeValue(network)
		if err != nil {
			t.Error(err)
		}
		newDoc, ok := v.(Document)
		if !ok {
			t.Errorf("Expected Document got %T", newDoc)
		}
		if !doc.Equal(newDoc) {
			t.Errorf("%#v did not equal %#v", newDoc, doc)
		}
	}
}

// to run benchmarks
// go test -v -bench Benchmark -run Benchmark -count 3

func BenchmarkEncodeBinary(b *testing.B) {
	buf := new(bytes.Buffer)
	for i := 0; i < b.N; i++ {
		buf.Reset()
		err := encodeDocument(buf, sampleLargeDoc, false)
		if err != nil {
			b.Error(err)
		}

	}
}

func BenchmarkEncodeJson(b *testing.B) {
	for i := 0; i < b.N; i++ {
		buf, err := json.Marshal(sampleLargeDoc)
		if err != nil {
			b.Error(err)
		}
		_ = buf
	}
}

func BenchmarkDecodeBinary(b *testing.B) {
	buf := new(bytes.Buffer)
	err := encodeDocument(buf, sampleLargeDoc, false)
	if err != nil {
		b.Error(err)
	}
	data := buf.Bytes()
	// reset timer
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = decodeValue(bytes.NewReader(data))
		if err != nil {
			b.Error(err)
		}

	}
}

func BenchmarkDecodeJson(b *testing.B) {
	buf, err := json.Marshal(sampleLargeDoc)
	if err != nil {
		b.Error(err)
	}
	// reset timer
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc := make(Document)
		err = json.Unmarshal(buf, &doc)
		if err != nil {
			b.Error(err)
		}
	}
}
