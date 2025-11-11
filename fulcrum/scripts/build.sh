
cp .env devfest2025.service install.sh ../output

cd ..
swag init

export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64
go build -ldflags="-s -w" -o output/devfest2025

