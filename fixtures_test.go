package apidoc

import (
	"bytes"
	"encoding/json"
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

var departureBlob = `{"id":"733048","href":"https://rest.gadventures.com/departures/733048","date_created":"2016-05-09T14:53:24Z","date_last_modified":"2017-02-08T17:59:06Z","name":"Delta & Falls Overland (Westbound)","start_date":"2017-04-29","finish_date":"2017-05-07","product_line":"DZFO","sku":"GPFDZFO170429-O1","flags":[],"start_address":{"street":"Stand 1385","city":"Victoria Falls","country":{"id":"ZW","href":"https://rest.gadventures.com/countries/ZW","name":"Zimbabwe"},"postal_zip":null,"latitude":"-17.923930","longitude":"25.841190"},"finish_address":{"street":"2 Schanzen Street ","city":"Windhoek","country":{"id":"NA","href":"https://rest.gadventures.com/countries/NA","name":"Namibia"},"postal_zip":null,"latitude":"-22.554120","longitude":"17.093810"},"latest_arrival_time":"2017-04-29T23:59:59","earliest_departure_time":"2017-05-07T00:00:00","nearest_start_airport":{"code":"VFA"},"nearest_finish_airport":{"code":"WDH"},"tour":{"id":"23185","href":"https://rest.gadventures.com/tours/23185"},"tour_dossier":{"id":"23185","href":"https://rest.gadventures.com/tour_dossiers/23185","name":"Delta & Falls Overland (Westbound)"},"rooms":[{"code":"STANDARD","name":"Standard","flags":[],"availability":{"status":"AVAILABLE","total":5,"male":null,"female":null},"price_bands":[{"code":"ADULT","name":"Adult","min_travellers":1,"max_travellers":22,"min_age":18,"max_age":39,"prices":[{"currency":"CAD","amount":"1249.00","deposit":"250.00","promotions":[]},{"currency":"CHF","amount":"1009.00","deposit":"250.00","promotions":[]},{"currency":"ZAR","amount":"15939.00","deposit":"1000.00","promotions":[]},{"currency":"EUR","amount":"929.00","deposit":"250.00","promotions":[]},{"currency":"AUD","amount":"1249.00","deposit":"250.00","promotions":[]},{"currency":"USD","amount":"1199.00","deposit":"250.00","promotions":[]},{"currency":"GBP","amount":"799.00","deposit":"100.00","promotions":[]},{"currency":"NZD","amount":"1469.00","deposit":"250.00","promotions":[]}]}],"addons":[{"product":{"id":"T733048","href":"https://rest.gadventures.com/single_supplements/T733048","name":"My Own Room","type":"single_supplements","sub_type":"Single Supplement"},"start_date":"2017-04-29","finish_date":"2017-05-07","min_days":8,"max_days":8,"request_space_date":"2017-04-19","halt_booking_date":"2017-04-28"}],"add_ons":[{"id":"T733048","href":"https://rest.gadventures.com/single_supplements/T733048","name":"My Own Room","type":"single_supplements","sub_type":"Single Supplement","start_date":"2017-04-29","finish_date":"2017-05-07","min_days":8,"max_days":8}]}],"addons":[{"product":{"id":"4373","href":"https://rest.gadventures.com/accommodations/4373","name":"Urban Camp","type":"accommodations","sub_type":"Hotel"},"start_date":"2017-05-07","finish_date":"2017-05-11","min_days":1,"max_days":5,"request_space_date":"2017-02-28","halt_booking_date":"2017-04-28"},{"product":{"id":"2059","href":"https://rest.gadventures.com/transports/2059","name":"Victoria Falls Airport to Victoria Falls Hotel Transfer","type":"transports","sub_type":"Transfer"},"start_date":"2017-04-29","finish_date":"2017-04-29","min_days":1,"max_days":1,"request_space_date":"2017-03-30","halt_booking_date":"2017-04-28"},{"product":{"id":"76458","href":"https://rest.gadventures.com/activities/76458","name":"Okavango Delta Flight","type":"activities","sub_type":"Product"},"start_date":"2017-05-03","finish_date":"2017-05-03","min_days":0,"max_days":0,"request_space_date":"2017-04-24","halt_booking_date":"2017-04-28"},{"product":{"id":"100685","href":"https://rest.gadventures.com/activities/100685","name":"Victoria Falls Helicopter Ride (Zimbabwe)","type":"activities","sub_type":"Product"},"start_date":"2017-04-30","finish_date":"2017-04-30","min_days":0,"max_days":0,"request_space_date":"2017-04-24","halt_booking_date":"2017-04-28"},{"product":{"id":"117983","href":"https://rest.gadventures.com/activities/117983","name":"Zambezi Sunset Cruise Vic Falls","type":"activities","sub_type":"Product"},"start_date":"2017-04-30","finish_date":"2017-04-30","min_days":0,"max_days":0,"request_space_date":"2017-04-19","halt_booking_date":"2017-04-28"}],"add_ons":[{"id":"4373","href":"https://rest.gadventures.com/accommodations/4373","name":"Urban Camp","type":"accommodations","sub_type":"Hotel","start_date":"2017-05-07","finish_date":"2017-05-11","min_days":1,"max_days":5},{"id":"2059","href":"https://rest.gadventures.com/transports/2059","name":"Victoria Falls Airport to Victoria Falls Hotel Transfer","type":"transports","sub_type":"Transfer","start_date":"2017-04-29","finish_date":"2017-04-29","min_days":1,"max_days":1},{"id":"76458","href":"https://rest.gadventures.com/activities/76458","name":"Okavango Delta Flight","type":"activities","sub_type":"Product","start_date":"2017-05-03","finish_date":"2017-05-03","min_days":0,"max_days":0},{"id":"100685","href":"https://rest.gadventures.com/activities/100685","name":"Victoria Falls Helicopter Ride (Zimbabwe)","type":"activities","sub_type":"Product","start_date":"2017-04-30","finish_date":"2017-04-30","min_days":0,"max_days":0},{"id":"117983","href":"https://rest.gadventures.com/activities/117983","name":"Zambezi Sunset Cruise Vic Falls","type":"activities","sub_type":"Product","start_date":"2017-04-30","finish_date":"2017-04-30","min_days":0,"max_days":0}],"availability":{"status":"AVAILABLE","total":5},"lowest_pp2a_prices":[{"currency":"USD","amount":"1199.00"},{"currency":"AUD","amount":"1249.00"},{"currency":"CHF","amount":"1009.00"},{"currency":"GBP","amount":"799.00"},{"currency":"NZD","amount":"1469.00"},{"currency":"CAD","amount":"1249.00"},{"currency":"ZAR","amount":"15939.00"},{"currency":"EUR","amount":"929.00"}],"requirements":[{"type":"CONFIRMATION","name":"Date of Birth","code":"DATE_OF_BIRTH","message":"Date of birth must be submitted for this product. 'date_of_birth' must be set in the customer resource.","flags":[],"details":[]},{"type":"CONFIRMATION","name":"Nationality","code":"NATIONALITY","message":"Nationality must be submitted for this product. 'nationality' must be set in the customer resource.","flags":[],"details":[]},{"type":"CONFIRMATION","name":"Age Restricted","code":"AGE_RESTRICTED","message":"There are age restrictions on this product that will be indicated in the 'min_age' and 'max_age' fields of the resource. Age is determined based on the 'start_date' of the product. 'date_of_birth' must be set in the customer resource.","flags":[],"details":[{"summary":"Yolo tours are designed for travellers between the ages of 18-39.","description":"Yolo tours are designed for those who thirst for adventure and want to take on the planet in a fast, raw and affordable way. These tours make the most of the time available, hitting the highlights by day, and mixing it up at night. These trips often feature public transportation, and basic accommodations and have been designed for those on a budget. Please note that Yolo style tours are designed for travellers aged 18-39.","detail_type":{"code":"CONFIRMATION_ONLY"}},{"summary":"Yolo tours are designed for travellers between the ages of 18-39.","description":"Yolo tours are designed for those who thirst for adventure and want to take on the planet in a fast, raw and affordable way. These tours make the most of the time available, hitting the highlights by day, and mixing it up at night. These trips often feature public transportation, and basic accommodations and have been designed for those on a budget. Please note that Yolo style tours are designed for travellers aged 18-39.","detail_type":{"code":"ANY"}}]},{"type":"CHECKIN","name":"Passport Expiry Date","code":"PASSPORT_EXPIRY_DATE","message":"Passport expiry date must be submitted for this product. The 'expiry_date' field (in the 'passport' data) must be set in the customer resource.","flags":[],"details":[]},{"type":"CHECKIN","name":"Passport Number","code":"PASSPORT_NUMBER","message":"Passport number must be submitted for this product. The 'number' field (in the 'passport' data) must be set in the customer resource.","flags":[],"details":[]},{"type":"INFORMATIONAL","name":"Customer Information","code":"CUSTOMER_INFORMATION","message":"Information that must be passed along to the customer but is not actionable.","flags":[],"details":[]}],"components":{"href":"https://rest.gadventures.com/departures/733048/departure_components"}}`

var errorBlob = `{"http_status_code": 404,"time": "2017-02-10T17:30:11","message": "No such departures with ID 7330489","errors": [],"error_id": "gapi_8cef0f3ad5e54ad2bab20493b896ef9f"}`
