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

type Target struct {
	Vendor    string
	Timestamp time.Time
	Point     geojson.Feature
}

func HandleHttpPost(nc *nats.Conn) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		vendorId := request.PathValue("vendorId")
		slog.Info("Received /vendor/{vendorId}/point POST request", "vendorId", vendorId)

		defer request.Body.Close()
		body, _ := io.ReadAll(request.Body)
		slog.Info("Processing request body", "body", string(body))

		f, err := validateAndParseFeature(body)
		slog.Info("Received feature", "type", f.Type, "geometry", f.Geometry, "props", f.Properties)

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

		target := Target{
			Point:     *f,
			Timestamp: time.Now(),
			Vendor:    vendorId,
		}

		data, _ := json.Marshal(target)

		msg := nats.NewMsg("msg.targets")
		msg.Header.Add("Nats-Msg-Id", uuid.NewString())
		msg.Data = data
		nc.PublishMsg(msg)

		encoder := json.NewEncoder(writer)
		encoder.Encode(target)
	}
}

func HandleTargetMessage(mongo *mongo.Client) func(jetstream.Msg) {
	return func(msg jetstream.Msg) {
		msgId := msg.Headers().Get("Nats-Msg-Id")
		slog.Info("Received a message from NATS stream", "ID", msgId)

		pointsCollection := mongo.Database("helloHttp").Collection("targets")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		t, err := validateAndParseTarget(msg.Data())

		if err != nil {
			slog.Error("Failed to parse GeoJSON message from stream", "ID", msgId, "message", string(msg.Data()))
			msg.Ack()
			return
		}

		res, err := pointsCollection.InsertOne(ctx, t)
		if err != nil {
			slog.Error("Failed to save target to the database", err)
			msg.Nak()
			return
		}

		slog.Info("Inserted a single document to MongoDB collection with id", "id", res.InsertedID)
		msg.Ack()
	}
}

func validateAndParseTarget(data []byte) (*Target, error) {
	target := &Target{}
	err := json.Unmarshal(data, &target)

	if err != nil {
		slog.Error("Error parsing Target", "error", err, "string", string(data))
	}

	return target, err
}

func validateAndParseFeature(data []byte) (*geojson.Feature, error) {
	f, err := geojson.UnmarshalFeature(data)

	// TODO: validate coordinates

	if err != nil {
		slog.Error("Error parsing GeoJSON", "error", err, "string", string(data))
	}

	return f, err
}
