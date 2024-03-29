## Testing
### Unit testing
For every push to any branch, GithubActions runs unit tests. You can run it locally using `go test $(go list ./... | grep -v integration)`. For integration tests, please see below.

There is a golang-ci linter as well. You can run it using `golangci-lint run`.

### Integration
Since I am a person from operation world, I prefer integration tests over unit testing.

The package [integration](../internal/integration) exists to keep all required test suites and scenarios. Some tests requires additional utilites and must be run in linux OS.

#### Examples
There are examples in [examples](../pulumi-template/examples) directory. Most of them used in GithubActions and tested for almost every commit to `main` branch.

The schema of example file name is
```
<k8s type>-<type of net>-<ha or non-ha>-<name>.yaml
```
Please follow this naming convention.
The `name` can be anything but do not make it too long, because there is a limitation in Hetzner server name.
