# golang-custom-body-router
Simple golang custom router.

### test server

```
# test uncompressed
curl -X POST -H "Content-Type: application/json" -d '{"key": "value"}' http://localhost:8080

# test compressed
curl -X POST -H "Content-Encoding: gzip" -H "Content-Type: application/json" --data-binary @compressed.json http://localhost:8080/
```