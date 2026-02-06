test TARGET:
  go test -C=./simulation -v -count=1 ./{{TARGET}}

build:
  go build -C=./simulation -o=./tmp/simulation ./cmd/main.go

start:
  ./simulation/tmp/simulation

update_deps:
  (cd ./simulation/ && go get -u ./... && go mod tidy)
