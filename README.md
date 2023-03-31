# Delta Storage Provider Selection

Simple filcoin storage provider selection.
- file size (min and max piece size)
- verified FIL (soon)
- geo location (soon)
- success rate throttle (soon)

## Running the app
```
go build -o simple-sp-selection
./simple-sp-selection
```

## Get a random SP within a given piece size range
```
curl --location 'http://localhost:8080/api/providers?size_bytes=256
```

## Test the live version
```
curl --location --request GET 'https://sp-select.delta.store/api/providers?size_bytes=256'
```
