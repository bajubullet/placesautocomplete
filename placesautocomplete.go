package placesautocomplete

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	apiSuggestionsURL = "https://maps.googleapis.com/maps/api/place/autocomplete/json"
	apiDetailURL      = "https://maps.googleapis.com/maps/api/place/details/json"
)

var (
	apiKey     = os.Getenv("GOOGLE_MAPS_API_KEY")
	httpClient = &http.Client{Timeout: 10 * time.Second}
)

type (
	StructuredAddress struct {
		MainText      string `json:"main_text"`
		SecondaryText string `json:"secondary_text"`
	}

	Prediction struct {
		PlaceID   string            `json:"place_id"`
		Addresses StructuredAddress `json:"structured_formatting"`
	}

	Suggessions struct {
		Predictions []Prediction
	}

	AddressComponent struct {
		Types []string
		Name  string `json:"long_name"`
	}

	RawPlaceDetail struct {
		Phone   string             `json:"international_phone_number"`
		Address []AddressComponent `json:"address_components"`
	}

	PlaceDetail struct {
		Phone      string
		Address    string
		PostalCode string
		City       string
		State      string
		Country    string
	}

	DetailResult struct {
		Result RawPlaceDetail
	}
)

func panic(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func getJSON(url string, data map[string]string, target interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	panic(err)
	q := req.URL.Query()
	for k, v := range data {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
	resp, err := httpClient.Do(req)
	panic(err)
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(target)
}

func GetSuggestions(q string) []Prediction {
	data := map[string]string{
		"key":    apiKey,
		"input":  q,
		"fields": "place_id,address_component", // Don't know why this is not working
	}
	result := new(Suggessions)
	getJSON(apiSuggestionsURL, data, result)
	return result.Predictions
}

func GetPlaceDetails(placeID string) PlaceDetail {
	data := map[string]string{
		"key":     apiKey,
		"placeid": placeID,
		"fields":  "address_component,international_phone_number",
	}
	result := new(DetailResult)
	getJSON(apiDetailURL, data, result)
	detail := new(PlaceDetail)
	detail.Phone = result.Result.Phone
	for _, address := range result.Result.Address {
		for _, addressType := range address.Types {
			switch addressType {
			case "postal_code":
				detail.PostalCode = address.Name
			case "country":
				detail.Country = address.Name
			case "administrative_area_level_1":
				detail.State = address.Name
			case "locality":
				detail.City = address.Name
			case "floor":
				detail.Address += " " + address.Name
			case "street_number":
				detail.Address += " " + address.Name
			case "route":
				detail.Address += " " + address.Name
			case "sublocality":
				detail.Address += " " + address.Name
			}
		}
	}
	detail.Address = strings.TrimSpace(detail.Address)
	return *detail
}
