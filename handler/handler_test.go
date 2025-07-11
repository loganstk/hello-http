package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kinbiko/jsonassert"
)

func TestPostHandler_OK(t *testing.T) {
	const requestBody = `{ "type": "FeatureCollection",
		"generator": "myapp",
		"timestamp": "2020-06-15T01:02:03Z",
		"features": [
		{ "type": "Feature",
			"geometry": {"type": "Point", "coordinates": [102.0, 0.5]},
			"properties": {"prop0": "value0"}
		}
		]
	}`

	bodyBytes := []byte(requestBody)

	req := httptest.NewRequest("POST", "/vendor/123/point", bytes.NewReader(bodyBytes))
	rr := httptest.NewRecorder()
	ja := jsonassert.New(t)

	HandlePostPoint(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	ja.Assertf(rr.Body.String(), requestBody)
}

func TestPostHandler_BadRequest(t *testing.T) {
	bodyBytes := []byte("not geojson")

	req := httptest.NewRequest("POST", "/vendor/123/point", bytes.NewReader(bodyBytes))
	rr := httptest.NewRecorder()
	ja := jsonassert.New(t)

	HandlePostPoint(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	const expected = `{"Code":1000,"Message":"Error parsing GeoJSON: invalid character 'o' in literal null (expecting 'u')"}`
	ja.Assertf(rr.Body.String(), expected)
}
