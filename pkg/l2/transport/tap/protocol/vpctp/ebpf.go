package vpctp

import (
	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/asm"
)

func vpctpEbpfProg(xsksMap *ebpf.Map, qidMap *ebpf.Map) (*ebpf.Program, error) {
	return ebpf.NewProgram(&ebpf.ProgramSpec{
		Name:          "vpc_xsk_ebpf",
		Type:          ebpf.XDP,
		License:       "LGPL-2.1 or BSD-2-Clause",
		KernelVersion: 0,
		Instructions: asm.Instructions{
			//# int xdp_sock_prog(struct xdp_md *ctx) {
			{OpCode: 0xbf, Src: 0x1, Dst: 0x6}, //0: r6 = r1
			//# int *qidconf, index = ctx->rx_queue_index;
			{OpCode: 0x61, Src: 0x6, Dst: 0x1, Offset: 0x10}, //1: r1 = *(u32 *)(r6 + 16)
			{OpCode: 0x63, Src: 0x1, Dst: 0xa, Offset: -4},   //2: *(u32 *)(r10 - 4) = r1
			{OpCode: 0xbf, Src: 0xa, Dst: 0x2},               //3: r2 = r10
			{OpCode: 0x07, Dst: 0x2, Constant: -4},           //4: r2 += -4
			//# qidconf = bpf_map_lookup_elem(&qidconf_map, &index);
			{OpCode: 0x18, Src: 0x1, Dst: 0x1, Constant: int64(qidMap.FD())}, //5: r1 = 0 ll
			{OpCode: 0x85, Constant: 0x1},                                    //7: call 1
			{OpCode: 0xbf, Src: 0x0, Dst: 0x1},                               //8: r1 = r0
			{OpCode: 0xb7, Dst: 0x0, Constant: 0},                            //9: r0 = 0
			//# if (!qidconf)
			{OpCode: 0x15, Dst: 0x1, Offset: 0x1f},   //10: if r1 == 0 goto +31
			{OpCode: 0xb7, Dst: 0x0, Constant: 0x02}, //11: r0 = 2
			//# if (!*qidconf)
			{OpCode: 0x61, Src: 0x1, Dst: 0x1, Constant: 0}, //12: r1 = *(u32 *)(r1 + 0)
			{OpCode: 0x15, Dst: 0x1, Offset: 0x1c},          //13: if r1 == 0 goto +28
			//# void *data_end = (void *)(long)ctx->data_end;
			{OpCode: 0x61, Src: 0x6, Dst: 0x2, Offset: 0x04}, //14: r2 = *(u32 *)(r6 + 4)
			//# void *data = (void *)(long)ctx->data;
			{OpCode: 0x61, Src: 0x6, Dst: 0x1, Constant: 0}, //15: r1 = *(u32 *)(r6 + 0)
			//# if (data + nh_off > data_end)
			{OpCode: 0xbf, Src: 0x1, Dst: 0x3},               //16: r3 = r1
			{OpCode: 0x07, Dst: 0x3, Constant: 0x0e},         //17: r3 += 14
			{OpCode: 0x2d, Src: 0x2, Dst: 0x3, Offset: 0x17}, //18: if r3 > r2 goto +23
			//# h_proto = eth->h_proto;
			{OpCode: 0x71, Src: 0x1, Dst: 0x4, Offset: 0x0c}, //19: r4 = *(u8 *)(r1 + 12)
			{OpCode: 0x71, Src: 0x1, Dst: 0x3, Offset: 0x0d}, //20: r3 = *(u8 *)(r1 + 13)
			{OpCode: 0x67, Dst: 0x3, Constant: 0x08},         //21: r3 <<= 8
			{OpCode: 0x4f, Src: 0x4, Dst: 0x3},               //22: r3 |= r4
			//# if (h_proto == __bpf_constant_htons(ETH_P_IP))
			{OpCode: 0x15, Dst: 0x3, Offset: 0x06, Constant: 0x86dd}, //23: if r3 == 56710 goto +6
			{OpCode: 0x55, Dst: 0x3, Offset: 0x11, Constant: 0x08},   //24: if r3 != 8 goto +17
			{OpCode: 0xb7, Dst: 0x3, Constant: 0x17},                 //25: r3 = 23
			//# if (iph + 1 > data_end)
			{OpCode: 0xbf, Src: 0x1, Dst: 0x4},               //26: r4 = r1
			{OpCode: 0x07, Dst: 0x4, Constant: 0x22},         //27: r4 += 34
			{OpCode: 0x2d, Src: 0x2, Dst: 0x4, Offset: 0x0d}, //28: if r4 > r2 goto +13
			{OpCode: 0x05, Offset: 0x04},                     //29: goto +4
			{OpCode: 0xb7, Dst: 0x3, Constant: 0x14},         //30: r3 = 20
			//# if (ip6h + 1 > data_end)
			{OpCode: 0xbf, Src: 0x1, Dst: 0x4},               //31: r4 = r1
			{OpCode: 0x07, Dst: 0x4, Constant: 0x36},         //32: r4 += 54
			{OpCode: 0x2d, Src: 0x2, Dst: 0x4, Offset: 0x08}, //33: if r4 > r2 goto +8
			{OpCode: 0x0f, Src: 0x3, Dst: 0x1},               //34: r1 += r3
			{OpCode: 0x71, Src: 0x1, Dst: 0x1, Constant: 0},  //35: r1 = *(u8 *)(r1 + 0)
			//# if (ipproto != VPC_PROTO)
			{OpCode: 0x55, Dst: 0x1, Offset: 0x05, Constant: IPPROTO}, //36: if r1 != 172 goto +5
			//# return bpf_redirect_map(&xsks_map, index, 0);
			{OpCode: 0x61, Src: 0xa, Dst: 0x2, Offset: -4},                    //37: r2 = *(u32 *)(r10 - 4)
			{OpCode: 0x18, Src: 0x1, Dst: 0x1, Constant: int64(xsksMap.FD())}, //38: r1 = 0 ll
			{OpCode: 0xb7, Dst: 0x3, Constant: 0},                             //40: r3 = 0
			{OpCode: 0x85, Constant: 0x33},                                    //41: call 51
			//# }
			{OpCode: 0x95}, //42: exit
		},
	})
}
