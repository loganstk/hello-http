package handler

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/paulmach/orb/geojson"
)

const (
	ErrInvalidJson = 1000
)

type ErrResponse struct {
	Code    int
	Message string
}

func PostHandler(nc *nats.Conn) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		slog.Info("Received /vendor/{vendorId} POST request", "vendorId", request.PathValue("vendorId"))

		defer request.Body.Close()
		body, _ := io.ReadAll(request.Body)
		slog.Info("Processing request body", "body", string(body))

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

		msg := nats.NewMsg("msg.points")
		msg.Header.Add("Nats-Msg-Id", uuid.NewString())
		msg.Data = body
		nc.PublishMsg(msg)

		for _, f := range fc.Features {
			slog.Info("Received feature", "type", f.Type, "geometry", f.Geometry, "props", f.Properties)
		}

		encoder := json.NewEncoder(writer)
		encoder.Encode(fc)
	}
}

func HandlePointMessage(mongo *mongo.Client) func(jetstream.Msg) {
	return func(msg jetstream.Msg) {
		msgId := msg.Headers().Get("Nats-Msg-Id")
		slog.Info("Received a message from NATS stream", "ID", msgId)

		pointsCollection := mongo.Database("helloHttp").Collection("points")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		fc, err := validateAndParse(msg.Data())

		if err != nil {
			slog.Error("Failed to parse GeoJSON message from stream", "ID", msgId, "message", string(msg.Data()))
			msg.Ack()
			return
		}

		res, err := pointsCollection.InsertOne(ctx, fc)
		if err != nil {
			slog.Error("Failed to save the point to the database", err)
			msg.Nak()
			return
		}

		slog.Info("Inserted a single document to MongoDB collection with id", "id", res.InsertedID)
		msg.Ack()
	}
}

func validateAndParse(data []byte) (*geojson.FeatureCollection, error) {
	fc, err := geojson.UnmarshalFeatureCollection(data)

	if err != nil {
		slog.Error("Error parsing GeoJSON", "error", err, "string", string(data))
	}

	return fc, err
}
