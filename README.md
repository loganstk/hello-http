# Golang + NATS + MongoDB helloworld app

1. Run app and services: `docker compose up --detach`.
2. Send POST requests to `http://localhost:8080/vendor/{{ vendorId }}/point` with sample payload:
```
{
  "type": "Feature",
  "geometry": {
    "type": "Point",
    "coordinates": [125.6, 10.1]
  },
  "properties": {
    "name": "Dinagat Islands"
  }
}
```

