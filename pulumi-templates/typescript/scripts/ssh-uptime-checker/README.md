# SSH checker
A Simple script for discovering rebooted node.

## Developing
Using docker is not recommended since containers shares the host kernel.

### Using vagrant
```
$ vagrant up
$ go run main.go 127.0.0.1 vagrant ./.vagrant/machines/default/virtualbox/private_key 2222
```
