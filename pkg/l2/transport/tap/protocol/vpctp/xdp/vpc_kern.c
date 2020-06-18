#define KBUILD_MODNAME "vpc_l2"
#include "types.h"
#include <uapi/linux/bpf.h>
#include <linux/bpf.h>
#include <asm/byteorder.h>
#include <linux/pkt_cls.h>
#include <linux/in.h>
#include <linux/if_ether.h>
#include <linux/if_packet.h>
#include <linux/ip.h>
#include <linux/ipv6.h>
#include "bpf_helpers.h"
#include "vpc_kern.h"

#define __bpf_constant_htons(x) ___constant_swab16(x)

//Ensure map references are available.
/*
	These will be initiated from go and 
	referenced in the end BPF opcodes by file descriptor
*/
struct {
	__uint(type, BPF_MAP_TYPE_XSKMAP);
	__type(key, int);
	__type(value, int);
	__uint(max_entries, MAX_SOCKS);
} xsks_map SEC(".maps");

struct {
	__uint(type, BPF_MAP_TYPE_ARRAY);
	__type(key, int);
	__type(value, int);
	__uint(max_entries, 2);
} qidconf_map SEC(".maps");

//Return the IPv4 next protocol
static int parse_ipv4(void *data, __u64 nh_off, void *data_end)
{
	struct iphdr *iph = data + nh_off;

	if (iph + 1 > data_end)
		return 0;

	return iph->protocol;
}

//Return the IPv6 next header (protocol)
static int parse_ipv6(void *data, __u64 nh_off, void *data_end)
{
	struct ipv6hdr *ip6h = data + nh_off;

	if (ip6h + 1 > data_end)
		return 0;

	return ip6h->nexthdr;
}

//Our main BPF code
SEC("xdp_sock")
int xdp_sock_prog(struct xdp_md *ctx) {
	int *qidconf, index = ctx->rx_queue_index;

	// A set entry here means that the correspnding queue_id
	// has an active AF_XDP socket bound to it.
	qidconf = bpf_map_lookup_elem(&qidconf_map, &index);
	if (!qidconf)
		return XDP_ABORTED;

	if (!*qidconf)
		return XDP_PASS;

	void *data_end = (void *)(long)ctx->data_end;
	void *data = (void *)(long)ctx->data;
	struct ethhdr *eth = data;

	__be16 h_proto;
	__u64 nh_off;
	__u32 ipproto;

	nh_off = sizeof(*eth);
	if (data + nh_off > data_end)
		return XDP_PASS;

	h_proto = eth->h_proto;

	//TODO(@tcfw) future support for VLAN tagging
	//Handle tagged packets
	// if (h_proto == __bpf_constant_htons(ETH_P_8021Q) ||
	// 	h_proto == __bpf_constant_htons(ETH_P_8021AD)) {
	// 	struct vlan_hdr *vhdr;

	// 	vhdr = data + nh_off;
	// 	nh_off += sizeof(struct vlan_hdr);
	// 	if (data + nh_off > data_end)
	// 		return XDP_PASS;
	// 	h_proto = vhdr->h_vlan_encapsulated_proto;
	// }

	//Get IP next protocol
	if (h_proto == __bpf_constant_htons(ETH_P_IP))
		ipproto = parse_ipv4(data, nh_off, data_end);
	else if (h_proto == __bpf_constant_htons(ETH_P_IPV6))
		ipproto = parse_ipv6(data, nh_off, data_end);
	else
		ipproto = 0;

	//Check if a VPC packet
	if (ipproto != VPC_PROTO) 
		return XDP_PASS;

	return bpf_redirect_map(&xsks_map, index, 0);
	// }
}

//Basic license just for compiling the object code
char __license[] SEC("license") = "LGPL-2.1 or BSD-2-Clause";