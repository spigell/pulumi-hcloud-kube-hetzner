## Usage

### Get ssh keys
```
pulumi stack output --show-secrets   -j ssh:keypair | jq .PrivateKey -r
```

### Get wg master key
```
pulumi stack output --show-secrets   wireguard:connection > ~/wg-dev.conf && wg-quick down ~/wg-dev.conf ; wg-quick up ~/wg-dev.conf
```

# RoadMap
## Code
- [ ] Rewrite wireguard part
- [ ] Rewrite ssh checker

## Tests
- [ ] Add idempotent tests for all runs
- [ ] Add tests for wireguard run (check master connection)
- [ ] Test with multiple servers
- [ ] Test with single node cluster (without leader tag)
