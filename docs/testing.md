## Testing

### Integration
Since I am a person from operation world, I prefer integration tests over unit testing.

The package [integration](../internal/integration) exists to keep all required test suites and scenarios. Some tests requires additional utilites and must be run in linux OS.

- wireguard_test: `wg-quick` cli and passwordless sudo (or run under root);

#### Examples
There are examples in [examples](../pulumi-template/examples) directory. Most of them used in GithubActions and tested for almost every commit to `main` branch.

The schema of example file name is
```
<k8s type>-<type of net>-<ha or non-ha>-<name>.yaml
```
Please follow this naming convention.
The `name` can be anything but do not make it too long, because there is a limitation in Hetzner server name.
