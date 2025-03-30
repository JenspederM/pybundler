# PyBundler

```sh
go run . bundle --output ./bundle -p ./examples/basic --overwrite

# Print help
./bundle/main --help

# Use basic entrypoint
./bundle/main basic
> Hello from basic!

# Use cli entrypoint
./bundle/main cli
> Hello from cli!
```