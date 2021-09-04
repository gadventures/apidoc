APIDoc [![Go Report Card](https://goreportcard.com/badge/github.com/gadventures/apidoc)](https://goreportcard.com/report/github.com/gadventures/apidoc) [![Go Reference](https://pkg.go.dev/badge/github.com/gadventures/apidoc.svg)](https://pkg.go.dev/github.com/gadventures/apidoc)
======

**WARNING:** alpha quality code; use at your own risk.

Generic `rest.gadventures.com (GAPI)` Resource datatype

It extracts the common code dealing with resources from GAPI across various projects.

Features include:

* Calculating checksums (ETag)
* Specific behaviour on how we de/serialize JSON (e.g. nil slices are [], not null)
* Equal method to be able to compare one resource to another
* Binary serialization (which is faster than JSON and thanks to golang/snappy, results in smaller size)
* Evaluating if response is a GAPIError

The core datatype `Document`, is an alias for `map[string]interface{}`. While
this will be slower than using custom structs, the tradeoff is that `Document`
instances do not require customization based on the requested resource. In
other words adding/removing attributes from GAPI resources does not require
modifications to the `gadventures/apidoc` module.

Usage
-----

```go
package main

import (
	"encoding/json"
	"log"

	"github.com/gadventures/apidoc"
)

const blob := `{
    "street": "Stand 1385",
    "city": "Victoria Falls",
    "country": {
        "id": "ZW",
        "href": "https://rest.gadventures.com/countries/ZW",
        "name": "Zimbabwe"
    }
}`

func main() {
	doc = apidoc.New()
	err := json.Unmarshal([]byte(blob), &doc)
	if err != nil {
		log.Fatal(err.Error())
	}
	
	fmt.Println(doc["city"].(string)) // will print Victoria Falls
	
	id, ok := doc.GetPath("country", "id")
	if !ok {
		log.Fatal("could not find country.id in document")
	}
	fmt.Println(countryID.(string)) // will print ZW
}
```
