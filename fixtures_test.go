package apidoc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func loadTestData(t *testing.T, filename string) []byte {
	t.Helper()

	data, err := os.ReadFile(fmt.Sprintf("testdata/%s", filename))
	if err != nil {
		t.Fatal(err)
	}

	return data
}

func TestDepartureBlob(t *testing.T) {
	var (
		doc, doc2 Document
		out       bytes.Buffer
	)
	buf := bytes.NewBuffer(loadTestData(t, "departure.json"))
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

	// count attributes
	var attribCount, attribCount2 int
	doc.TraverseCall(getAttributeCounts(&attribCount))
	delete(doc, "truth")
	doc.TraverseCall(getAttributeCounts(&attribCount2))
	if attribCount-1 != attribCount2 {
		t.Errorf("Expected the two counts to equal %d %d",
			attribCount, attribCount2)
	}
}

func TestDepartureCopy(t *testing.T) {
	departureBlob := loadTestData(t, "departure.json")

	var doc, doc2 Document

	err := json.NewDecoder(bytes.NewBuffer(departureBlob)).Decode(&doc)
	if err != nil {
		t.Error(err)
	}
	doc2 = *doc.Copy()
	if !doc.Equal(doc2) {
		t.Errorf("The two documents must equal")
	}
}

func TestDepartureToGAPIErrorNil(t *testing.T) {
	departureBlob := loadTestData(t, "departure.json")

	var doc Document
	err := json.NewDecoder(bytes.NewBuffer(departureBlob)).Decode(&doc)
	if err != nil {
		t.Error(err)
	}

	// expecting a nil GAPIError from a valid Document
	nilGAPIErr := doc.GAPIError(doc["href"].(string))
	if nilGAPIErr != nil {
		t.Errorf("valid Document should not return a GAPIError")
	}
}

func getAttributeCounts(count *int) TraverseFunc {
	return func(doc *Document, attributeName string, attributeValue interface{}) {
		*count++
	}
}
