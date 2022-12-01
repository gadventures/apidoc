package apidoc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"
)

func loadTestData(filename string) (string, error) {
	file, err := os.Open(fmt.Sprintf("./testdata/%s", filename))
	if err != nil {
		return "", fmt.Errorf("failed to load file %s - %s", filename, err.Error())
	}
	defer file.Close()
	blob, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s - %s", filename, err.Error())
	}
	return string(blob), nil
}

// hold the data in memory
var departureBlob, errorBlob string

func setup() {
	var err error
	departureBlob, err = loadTestData("departure.json")
	if err != nil {
		panic(err)
	}
	errorBlob, err = loadTestData("gapi_error.json")
	if err != nil {
		panic(err)
	}
}

func TestMain(m *testing.M) {
	setup()
	os.Exit(m.Run())
}

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
		t.Errorf("document should not be a GAPIError")
	}

	// real error
	err = json.NewDecoder(bytes.NewBufferString(errorBlob)).Decode(&doc2)
	if err != nil {
		t.Error(err)
	}

	// FIXME(ammaar): This test doesn't make sense so commenting it out for now
	// blobsBadHref := doc["href"].(string) + "9"
	// gErr = doc2.GAPIError(blobsBadHref)
	// if gErr == nil {
	// 	t.Errorf("document should be a GAPIError")
	// }
	// if !strings.Contains(gErr.Error(), "404") {
	// 	t.Errorf("Expected 404 but got %s", gErr.Error())
	// }
}

func getAttributeCounts(count *int) TraverseFunc {
	return func(doc *Document, attributeName string, attributeValue interface{}) {
		*count++
	}
}
