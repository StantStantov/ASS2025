build:
  go build -C=./simulation -o=./tmp/simulation ./cmd/main.go

start:
  ./simulation/tmp/simulation
