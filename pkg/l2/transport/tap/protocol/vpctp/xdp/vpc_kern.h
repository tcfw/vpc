#define VPC_PROTO 172
#define MAX_SOCKS 16

struct vlan_hdr {
	__be16 h_vlan_TCI;
	__be16 h_vlan_encapsulated_proto;
};