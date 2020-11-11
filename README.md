<div style="text-align: center;">
<h1>VPC</h1>
Create simple VxLAN based VPC's written in (mostly) Go using Linux bridges/netlink, iptables & network namespaces.
<p>
<a href="https://goreportcard.com/report/github.com/tcfw/vpc" title="Go Report Card"><img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/tcfw/vpc"></a>
</p>
</div>
<hr/>

## Why?
For fun and to learn!

# Schematic
![vpc](./docs/res/vpc.jpg "VPC")

## Host Types
 - Compute hosts facilitate the creation and management of VM's or containers 
 - Route hosts provide virtual routers
 - Management hosts (not in diagram) provide management facilities such as BGP route reflection and configuration management. 

> There is no technical reason why a compute host can also be a route host and vice-versa. This simply provides better security, bandwidth and segregation of duties on hosts.

## Connectivity
Each VM, namespace or container is connected to a Linux bridge (with VLAN filtering enabled) on a compute host. Compute hosts are connected via Linux VxLAN devices (VTEPs).

## VTEP Learning
VxLAN learning is disabled by default. Learning is derived from an ML-BGP-L2VPN-EVPN client (via [frr](https://github.com/FRRouting/frr)) on each compute host and route reflectors on management hosts. 

## Segregation
Each 'tenant' is separated by VxLAN VNI's and each Subnet is protected via inner VLAN tagging on a Linux bridge per tenant. 

# Agents
## L2
The L2 agent provides a GRPC API to create bridges, VxLAN VTEPs and manage VLAN tagging on the bridges.

### Transports
Can set up to use a linux VxLAN device, or use a TAP device with VxLAN encapsulation. The TAP device allows easier handling of ARP/ICMPv6 soliciations in the future.

## L3
The L3 agent provides the functionality to create the virtual router namespaces and provide simple DHCP/NAT & routing capabilities.

## SBS
Simple block storage - raft based replicated block storage medium exposing NBD endpoints

# Similar architectures
[Openstacks Neutron](https://wiki.openstack.org/wiki/Neutron) in Linux bridge mode.
