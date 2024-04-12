## Testing
### Unit testing
For every push to any branch, GithubActions runs unit tests. You can run them locally using `go test $(go list ./... | grep -v integration)`. For integration tests, please see below.

There is a golang-ci linter as well. You can run it using `golangci-lint run`.

### Integration
Since I am a person from the operations world, I prefer integration tests over unit testing.

The package [integration](../internal/integration) exists to keep all required test suites and scenarios. Some tests require additional utilities and must be run in the Linux OS.

#### Examples
There are examples in the [examples](../pulumi-template/examples) directory. Most of them are used in GithubActions and tested for almost every commit to the `main` branch.

The schema of the example file name is:
```
<k8s type>-<type of net>-<ha or non-ha>-<name>.yaml
```
Please follow this naming convention.
The `name` can be anything but do not make it too long, because there is a limitation in the Hetzner server name.
