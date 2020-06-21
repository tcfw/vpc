# Example
Running GoBGP in standalone `(sudo) gobgpd -f ./reflector.yaml`
```
global:
  config:
    as: 65000
    router-id: 192.168.122.1
    local-address-list:
      - 192.168.122.1
peer-groups:
  - config:
      peer-group-name: l2vpn
    afi-safis:
      - config:
          afi-safi-name: l2vpn-evpn
    route-reflector:
      config:
        route-reflector-client: true
        route-reflector-cluster-id: 192.168.122.1
dynamic-neighbors:
  - config:
      peer-group: l2vpn
      prefix: 192.168.122.0/24
```