package apidoc

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestDepartureBlob(t *testing.T) {
	var (
		doc, doc2 Document
		out       bytes.Buffer
	)
	buf := bytes.NewBufferString(departureBlob)
	err := json.NewDecoder(buf).Decode(&doc)
	if err != nil {
		t.Error(err)
	}
	doc["truth"] = true
	doc["lie"] = false
	err = json.NewEncoder(&out).Encode(doc)
	if err != nil {
		t.Error(err)
	}
	err = json.NewDecoder(&out).Decode(&doc2)
	if err != nil {
		t.Error(err)
	}
	if !doc.Equal(doc2) {
		t.Errorf("The two documents must equal")
	}
	if doc.String() != doc2.String() {
		t.Errorf("The two documents string must equal")
	}

	// get path check
	val, ok := doc.GetPath("start_address", "country", "name")
	if !ok {
		t.Errorf("Expected to found the country name")
	}
	expected := "Zimbabwe"
	if val != expected {
		t.Errorf("Expected %s but got %s", expected, val)
	}

	// count attribs
	var attribCount, attribCount2 int
	doc.TraverseCall(getCountAttribs(&attribCount))
	delete(doc, "truth")
	doc.TraverseCall(getCountAttribs(&attribCount2))
	if attribCount-1 != attribCount2 {
		t.Errorf("Expected the two counts to equal %d %d",
			attribCount, attribCount2)
	}
}

func TestDepartureAdditional(t *testing.T) {
	var doc, doc2 Document
	err := json.NewDecoder(bytes.NewBufferString(departureBlob)).Decode(&doc)
	if err != nil {
		t.Error(err)
	}
	doc2 = *doc.Copy()
	if !doc.Equal(doc2) {
		t.Errorf("The two documents must equal")
	}

	// nil error
	gErr := doc2.GAPIError(doc2["href"].(string))
	if gErr != nil {
		t.Errorf("document should not be an GAPIError")
	}

	// real error
	err = json.NewDecoder(bytes.NewBufferString(errorBlob)).Decode(&doc2)
	if err != nil {
		t.Error(err)
	}

	blobsBadHref := doc["href"].(string) + "9"
	gErr = doc2.GAPIError(blobsBadHref)
	if gErr == nil {
		t.Errorf("document should not an GAPIError")
	}
	if !strings.Contains(gErr.Error(), "404") {
		t.Errorf("Expected 404 but got %s", gErr.Error())
	}
}

func getCountAttribs(count *int) TraverseFunc {
	return func(doc *Document, attributeName string, attributeValue interface{}) {
		*count++
	}
}
