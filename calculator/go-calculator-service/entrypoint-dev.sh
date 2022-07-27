# If DEVELOPMENT_SEARCH_PATH environment variable specified, search for go-inventa package 
# in the specified path. This required for development of go-inventa itself.
if [ ! -z "$DEVELOPMENT_SEARCH_PATH" ];then
    export GOWORK="$PWD/go_dev.work"
fi

echo "Working with GOWORK=$GOWORK..."

echo "Downloading dependent Go modules..."
go mod download -x
echo "Running into Waiting loop..."
tail -f /dev/null