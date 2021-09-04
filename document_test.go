package apidoc

import (
	"testing"
)

func TestDocumentEqual(t *testing.T) {
	doc1 := make(Document)
	doc2 := make(Document)
	doc3 := make(Document)
	doc1["a"] = "Adam"
	doc1["b"] = "Boris"
	doc2["b"] = "Boris"
	doc2["a"] = "Adam"
	doc3["a"] = "Adam"
	doc3["b"] = "Basil"
	equals(t, 2, len(doc1))
	assert(t, doc1.Equal(doc2), "doc1 == doc2")
	assert(t, doc2.Equal(doc1), "doc2 == doc1")
	assert(t, !doc2.Equal(doc3), "doc2 != doc3")
	doc2["c"] = 3.0
	assert(t, !doc1.Equal(doc2), "doc1 != doc2")
	doc1["d"] = 3.0
	assert(t, !doc1.Equal(doc2), "doc1 != doc2")
	doc1["c"] = 3.0
	doc2["d"] = doc3
	doc1["d"] = doc3
	assert(t, doc1.Equal(doc2), "doc1 == doc2")
	// doc4
	doc4 := make(Document)
	doc4["a"] = "Adam"
	doc4["b"] = "Basil"
	doc1["d"] = doc4
	assert(t, doc1.Equal(doc2), "doc1 == doc2")
	doc4["b"] = "Ben"
	assert(t, !doc1.Equal(doc2), "doc1 != doc2")
	doc4["b"] = "Basil"
	assert(t, doc1.Equal(doc2), "doc1 == doc2")
	doc2["d"] = 5.6
	assert(t, !doc1.Equal(doc2), "doc1 != doc2")
	doc2["d"] = doc3
	assert(t, doc1.Equal(doc2), "doc1 == doc2")
	// floats
	doc5 := make(Document)
	doc6 := make(Document)
	doc5["foo"] = float64(5.4)
	doc6["foo"] = float64(5.4)
	assert(t, doc5.Equal(doc6), "doc5 == doc6")
	assert(t, doc6.Equal(doc5), "doc5 == doc6")
	doc6["foo"] = float64(5.41)
	assert(t, !doc5.Equal(doc6), "doc5 != doc6")
	assert(t, !doc6.Equal(doc5), "doc5 != doc6")
	// invalids int have no place there
	doc5["foo"] = float64(5)
	doc6["foo"] = float64(5)
	assert(t, doc5.Equal(doc6), "doc5 == doc6")
	assert(t, doc6.Equal(doc5), "doc5 == doc6")
	doc6["foo"] = int(5)
	assert(t, !doc5.Equal(doc6), "doc5 != doc6")
	assert(t, !doc6.Equal(doc5), "doc5 != doc6")
}

func TestDocumentEqualWithNils(t *testing.T) {
	makeDocument := func() Document {
		doc := make(Document)
		doc["a"] = nil
		doc["b"] = nil
		doc["c"] = nil
		doc["d"] = nil
		doc["f"] = nil
		doc["g"] = nil
		doc["h"] = "Hi"
		doc["h1"] = float64(4.5)
		doc["h2"] = true
		return doc
	}
	doc1 := makeDocument()
	doc2 := makeDocument()
	assert(t, doc1.Equal(doc2), "doc1 == doc2 fail")
	doc2["h"] = "Hm"
	assert(t, !doc1.Equal(doc2), "doc1 != doc2 fail")
	assert(t, !doc2.Equal(doc1), "doc1 != doc2 fail")
	// from doc2 point of view
	assert(t, !doc2.Equal(doc1), "doc1 != doc2 fail")
	assert(t, !doc1.Equal(doc2), "doc1 != doc2 fail")
	doc2["h"] = "Hi"
	assert(t, doc2.Equal(doc1), "doc2 == doc1 fail")
	assert(t, doc1.Equal(doc2), "doc2 == doc1 fail")
	// ok now add one extra attrib to doc2
	doc2["extra"] = "extra"
	assert(t, !doc1.Equal(doc2), "doc1 != doc2 fail")
	assert(t, !doc2.Equal(doc1), "doc1 != doc2 fail")
}

func TestDocumentEqualSlices(t *testing.T) {
	// test the slices of Document
	makeDocument := func() Document {
		doc := make(Document)
		doc["h"] = "Hi"
		m := make(Document)
		m["foo"] = "bar"
		slc := make([]interface{}, 10)
		for i := 0; i < len(slc); i++ {
			slc[i] = m
		}
		doc["s"] = slc
		return doc
	}
	doc1 := makeDocument()
	doc2 := makeDocument()
	assert(t, doc1.Equal(doc2), "doc1 == doc2 fail")
	assert(t, doc2.Equal(doc1), "doc1 == doc2 fail")
	// now mess with doc2 slice some
	fd := make(Document)
	fd["foo"] = "bar"
	// ensure fd equivalent to what the doc holds
	mslcif, prs := doc2["s"]
	assert(t, prs, "key presence fail")
	mslc, tpok := mslcif.([]interface{})
	assert(t, tpok, "cast fail")
	// validate
	assert(t, mslc[0].(Document).Equal(fd), "doc == map fail")
	assert(t, fd.Equal(mslc[0].(Document)), "doc == map fail")
	// do new slice of maps
	nslc := make([]interface{}, 10)
	for i := 0; i < len(nslc); i++ {
		nslc[i] = fd
	}
	doc2["s"] = nslc
	assert(t, doc1.Equal(doc2), "doc1 == doc2 fail")
	assert(t, doc2.Equal(doc1), "doc1 == doc2 fail")
	// diff sizes check
	doc2["s"] = nslc[:3]
	assert(t, !doc1.Equal(doc2), "doc1 != doc2 fail")
	assert(t, !doc2.Equal(doc1), "doc1 != doc2 fail")
	// reset
	doc2["s"] = nslc
	assert(t, doc1.Equal(doc2), "doc1 == doc2 fail")
	assert(t, doc2.Equal(doc1), "doc1 == doc2 fail")
	// change a value check
	nslc[3] = false
	doc2["s"] = nslc
	assert(t, !doc1.Equal(doc2), "doc1 != doc2 fail")
	assert(t, !doc2.Equal(doc1), "doc1 != doc2 fail")
	// nil slice
	doc2["s"] = nil
	assert(t, !doc1.Equal(doc2), "doc1 != doc2 fail")
	assert(t, !doc2.Equal(doc1), "doc1 != doc2 fail")
}

func TestDocumentETags(t *testing.T) {
	equalsTag := func(tag ETag, doc Document) bool {
		nt, err := doc.ETag()
		if err != nil {
			t.Error(err)
		}
		return tag == nt
	}
	doc := New()
	doc["id"] = float64(13)
	doc["zame"] = "foo"
	doc["yame"] = "foo"
	doc["same"] = "foo"
	doc["rame"] = "foo"
	doc["qame"] = "foo"
	doc["pame"] = "foo"
	doc["uame"] = "foo"
	doc["xame"] = "foo"
	doc["vame"] = "foo"
	doc["aame"] = "foo"
	doc["eame"] = "foo"
	doc["jame"] = "foo"
	doc["kame"] = "foo"
	doc["name"] = "foo"
	doc["embed"] = *(doc.Copy())
	// plain diff
	tag, err := doc.ETag()
	ok(t, err)
	assert(t, uint64(tag) > 0, "tag more than zero")
	doc2 := *(doc.Copy())
	assert(t, equalsTag(tag, doc2), "tags should equal")
	doc2["name"] = "foobar"
	assert(t, !equalsTag(tag, doc2), "doc2 should be diff")
	doc2["name"] = "foo"
	assert(t, equalsTag(tag, doc2), "should equal")
	// since we are using json to compute the hash this also confirm repeatability on both json and by extension binary encodings (as BinaryMarshaller uses json internaly as well)
	for i := 0; i < 10000; i++ {
		doc3 := *(doc.Copy())
		doc3["foo"] = "bar"
		doc3["name"] = "foo"
		assert(t, !equalsTag(tag, doc3), "doc3 should be diff")
		delete(doc3, "foo")
		d3tag, err := doc3.ETag()
		ok(t, err)
		equals(t, tag, d3tag)
		equals(t, tag.String(), d3tag.String())
	}
	d2tag, err := doc2.ETag()
	ok(t, err)
	newTag, err := NewETag(d2tag.String())
	ok(t, err)
	equals(t, tag, newTag)
	badTag, err := NewETag("z")
	assert(t, err != nil, "expected err")
	equals(t, uint64(0), uint64(badTag))
}

func TestDocumentBinary(t *testing.T) {
	doc := New()
	doc["id"] = float64(13)
	doc["zame"] = "foo"
	doc["yame"] = "foo"
	doc["same"] = "foo"
	doc["rame"] = "foo"
	doc["qame"] = "foo"
	doc["pame"] = "foo"
	doc["uame"] = "foo"
	doc["xame"] = "foo"
	doc["vame"] = "foo"
	doc["aame"] = float64(15.0)
	doc["eame"] = "foo"
	doc["jame"] = "foo"
	doc["kame"] = "foo"
	doc["name"] = "foo"
	doc["embed"] = *(doc.Copy())
	serial, err := doc.MarshalBinary()
	ok(t, err)
	var doc1 Document
	err = doc1.UnmarshalBinary(serial)
	ok(t, err)
	tag, err := doc1.ETag()
	ok(t, err)
	doctag, err := doc.ETag()
	ok(t, err)
	equals(t, doctag, tag)
	doc["id"] = "boo"
	doctag, err = doc.ETag()
	ok(t, err)
	assert(t, doctag != tag, "docs are diff now")
	doc["id"] = float64(13)
	doctag, err = doc.ETag()
	ok(t, err)
	equals(t, doctag, tag)
	// should be good
}
