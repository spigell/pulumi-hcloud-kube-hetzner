## Testing

Since I am a person from operation world, I prefer integration testing over unit testing.

There are examples in [examples](../pulumi-template/examples) directory. Most of them used in GithubActions and tested for almost every commit to `main` branch.

The schema of example file name is 
```
<k8s type>-<type of net>-<ha or non-ha>-<name>.yaml
```
Please follow this naming convention.
The `name` can be anything but do not make it too long, because there is a limitation in Hetzner server name.
