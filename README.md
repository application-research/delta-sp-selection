# Delta Storage Provide Selection

Simple filcoin storage provider selection.

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
curl --location --request GET 'https://simple-sp-selection.onrender.com/api/providers?min_piece_size_bytes=0&max_piece_size_bytes=34359738368'
```
