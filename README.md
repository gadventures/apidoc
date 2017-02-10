### APIDoc

[![GoDoc](https://godoc.org/github.com/gadventures/apidoc?status.svg)](https://godoc.org/github.com/gadventures/apidoc)   

Generic rest.gadventures.com resource datatype

A lot of code when dealing with resources from rest.gadventures.com is common between various projects.

This includes
* Calculating checksums(ETag)
* Specific behaviour on how we serialize to/from JSON (nil slices are [] not null)
* Equal method to be able to compare one resource to another
* Binary serialization (which is both faster than JSON and thanks to snappy being used under hood results in smaller size as well)
* Evaluating if response is an GAPIError

This datatype is basically alias for map[string]interface{}.
This will be slower than using custom structs.
The tradeoff is that APIDoc instances do not require specific customization based on the resource used.
In other words adding/removing attributes from rest.gadventures.com resources does not require modifications to APIDoc.

Usage:

```golang
import "encoding/json"

import "github.com/gadventures/apidoc"

blob := `
{                                                                                                                                                                                                                   
"street": "Stand 1385",                                                                                                                                                                                                            
"city": "Victoria Falls",                                                                                                                                                                                                          
"country": {                                                                                                                                                                                                                       
  "id": "ZW",                                                                                                                                                                                                                      
  "href": "https://rest.gadventures.com/countries/ZW",                                                                                                                                                                             
  "name": "Zimbabwe"                                                                                                                                                                                                               
  }
}
`

doc = apidoc.New()
err := json.Unmarshal([]byte(blob), &doc)

fmt.Println(doc["city"].(string)) //will print Victoria Falls

countryID, ok := doc.GetPath("country", "id")
if !ok {
	// oops did not find country id
}
fmt.Println(countryID.(string)) //will print ZW


```

Extracted and published from bundler project.

Alpha quality code - use at your own risk.
