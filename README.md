# Simple SP Selection

Simple filcoin storage provider selection.

## Running the app
```
go build -o simple-sp-selection
./simple-sp-selection
```

## Test the live one here

### Get a random SP within a given piece size range
```
curl --location 'http://localhost:8080/api/providers?size_bytes=256
```

