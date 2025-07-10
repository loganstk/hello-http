# Golang + NATS + MongoDB helloworld app

1. Run backing services: `docker compose up --detach`.
2. Send POST requests to `http://localhost:8080/vendor/{{ vendorId }}/point`. Sample payload:
```{ 
  "type": "FeatureCollection",
  "features": [
    { "type": "Feature",
      "geometry": {"type": "Point", "coordinates": [102.0, 0.5]},
      "properties": {"prop0": "value0"}
    }
  ]
}```

