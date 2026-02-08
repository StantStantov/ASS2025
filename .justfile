test TARGET:
  go test -C=./simulation -v -count=1 ./{{TARGET}}

bench COUNT TARGET:
  go test -C=./simulation -v -bench=. -benchmem -benchtime=100000x -count={{COUNT}} -cpu=8 ./{{TARGET}}

build:
  go build -C=./simulation -o=./tmp/simulation ./cmd/main.go

start:
  ./simulation/tmp/simulation

update_deps:
  (cd ./simulation/ && go get -u ./... && go mod tidy)
