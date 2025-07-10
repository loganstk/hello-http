package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/loganstk/hello-http/handler"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if len(os.Args) < 2 {
		slog.Info("Please, specify operation mode. Possible values are [publisher|consumer]")
		os.Exit(0)
	}

	nc, err := nats.Connect(os.Getenv("NATS_URL"))

	if err != nil {
		slog.Error("Error connecting to NATS server", "NATS_URL", os.Getenv("NATS_URL"))
		return
	}

	defer nc.Drain()

	switch os.Args[1] {
	case "publisher":
		slog.Info("Starting HTTP server...")
		http.HandleFunc("POST /vendor/{vendorId}/point", handler.PostHandler(nc))
		http.ListenAndServe(":8080", nil)
	case "consumer":
		slog.Info("Starting NATS stream consumer...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		js, _ := jetstream.New(nc)

		str, _ := js.CreateStream(ctx, jetstream.StreamConfig{
			Name:     "points",
			Subjects: []string{"msg.points"},
		})

		mongo, _ := mongo.Connect(options.Client().ApplyURI(os.Getenv("MONGODB_URL")))

		cons, _ := str.CreateConsumer(ctx, jetstream.ConsumerConfig{
			Durable:   "points",
			AckPolicy: jetstream.AckExplicitPolicy,
		})

		cc, err := cons.Consume(handler.HandlePointMessage(mongo))

		if err != nil {
			slog.Error("Error creating NATS stream consumer", err)
		}

		select {
		case <-ctx.Done():
			cc.Stop()
		}
	default:
		slog.Info("Possible values are [publisher|consumer]")
		os.Exit(0)
	}
}
