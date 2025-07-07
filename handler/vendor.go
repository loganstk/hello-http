package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/paulmach/orb/geojson"
)

const (
	ErrInvalidJson = 1000
)

type ErrResponse struct {
	Code    int
	Message string
}

func PostHandler(writer http.ResponseWriter, request *http.Request) {
	vendorId := request.PathValue("vendorId")
	log.Printf("Vendor ID path param: %v\n", vendorId)

	defer request.Body.Close()
	body, _ := io.ReadAll(request.Body)
	log.Printf("Request body: %s\n", string(body))

	fc, err := validateAndParse(body)

	if err != nil {
		http.ResponseWriter.WriteHeader(writer, http.StatusBadRequest)

		responseObj := ErrResponse{
			ErrInvalidJson,
			err.Error(),
		}

		encoder := json.NewEncoder(writer)
		encoder.Encode(responseObj)

		return
	}

	for _, f := range fc.Features {
		log.Printf("Feature type: %v\n", f.Type)
		log.Printf("Feature type: %v\n", f.Geometry)
		log.Printf("Feature type: %v\n", f.Properties)
	}

	encoder := json.NewEncoder(writer)
	encoder.Encode(fc)
}

func validateAndParse(data []byte) (*geojson.FeatureCollection, error) {
	fc, err := geojson.UnmarshalFeatureCollection(data)

	if err != nil {
		errorMsg := fmt.Sprintf("Error parsing GeoJSON: %v", err)
		log.Println(errorMsg)
		err = errors.New(errorMsg)
	}

	return fc, err
}
