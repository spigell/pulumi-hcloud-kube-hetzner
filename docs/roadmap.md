# RoadMap
## Documentation
- [ ] Add doc generation from structs
- [ ] Describe network modes
- [ ] Describe project layout
- [ ] Spelling
- [ ] Add roadmap for autoscaling
- [ ] Add readme for SDK

## Features
### pre-0.1.0
- [ ] Add support for "destroyed" cluster
- [ ] Add a licence
- [ ] Add basic k8s apps (VM, hetzner MCC, upgrader, kured)
- [ ] Add state for internal network IPAM
- [ ] Replace the golang script for ssh checking with binary
- [ ] Add dotnet and python SDK
- [ ] Add more validation rules (size of the net, difference between servers flags)
- [x] Rewrite ssh checker
- [x] Error checking for systemctl services
- [x] Set timeouts for Command resources
- [x] Expose kubeApiServer endpoint
- [x] Expose kubeconfig
- [x] Add a external ip of the program to FW rules
- [x] Add auto upgrade management for microos
- [x] Add dynamic version detection
- [x] Add an ability to run cluster without leader tag with single master
- [x] K3s token generation
- [x] Add fw rules for the public network mode
- [x] Mark all sensitive values as secrets

### after-0.1.0
- [x] Add reasonable defaults for variables
- [ ] Add autoscaling
- [ ] Add arm64 support
- [ ] Allow change config from code
- [ ] Package stage: reboot if changes detected only

## CI
- [ ] Find a way to hide output of the pulumi command plugin for several stages
- [x] Add linter run for every branch
- [x] Add go test run for every branch
- [x] Use pulumi cli instead of actions for up and preview. Collect logs.

## Tests
- [x] Test with multiple servers
- [x] Test with single node cluster (without leader tag)
