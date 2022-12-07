# gadventures/apidoc

[![Go Report Card](https://goreportcard.com/badge/github.com/gadventures/apidoc)](https://goreportcard.com/report/github.com/gadventures/apidoc) [![Go Reference](https://pkg.go.dev/badge/github.com/gadventures/apidoc.svg)](https://pkg.go.dev/github.com/gadventures/apidoc)

**WARNING:** alpha quality code; use at your own risk.

Generic <rest.gadventures.com> G API `Document` datatype

It extracts the common code for dealing with resource (JSON) "documents" from
the G API.

## Features include

* Calculating checksums (`ETag`)
* Specific behaviour on how the G API de/serializes JSON (e.g. nil slices are [], not null)
* `Equal` method to be able to compare one `Document` to another
* Binary serialization, which is faster than JSON and thanks to `golang/snappy`, results in smaller payload size
* Evaluating if the response is a `GAPIError`

The core datatype `Document`, is an alias for `map[string]interface{}`. While
this will be slower than using custom structs, the tradeoff is that `Document`
instances do not require customization based on the requested resource. In
other words adding/removing attributes from G API resources does not require
modifications to the `gadventures/apidoc` module.

## Usage

specify `require github.com/gadventures/apidoc v0.2.0` in your `go.mod` file

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
	
	fmt.Println(doc["city"].(string)) // will print "Victoria Falls"
	
	id, ok := doc.GetPath("country", "id")
	if !ok {
		log.Fatal("could not find country.id in document")
	}
	fmt.Println(countryID.(string)) // will print "ZW"
}
```
