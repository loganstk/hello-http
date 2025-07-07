package handler

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/paulmach/orb/geojson"
)

const (
	ErrInvalidJson = 1000
)

type ErrResponse struct {
	Code    int
	Message string
}

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func PostHandler(writer http.ResponseWriter, request *http.Request) {

	logger.Info("Received /vendor/{vendorId} POST request", "vendorId", request.PathValue("vendorId"))

	defer request.Body.Close()
	body, _ := io.ReadAll(request.Body)
	logger.Info("Processing request body", "body", string(body))

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
		logger.Info("Received feature", "type", f.Type, "geometry", f.Geometry, "props", f.Properties)
	}

	encoder := json.NewEncoder(writer)
	encoder.Encode(fc)
}

func validateAndParse(data []byte) (*geojson.FeatureCollection, error) {
	fc, err := geojson.UnmarshalFeatureCollection(data)

	if err != nil {
		logger.Error("Error parsing GeoJSON", "error", err, "string", string(data))
	}

	return fc, err
}
