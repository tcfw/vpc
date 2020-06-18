# Converting object files to BPF ASM
Usually you can just load the object file into the network interface by using iproute2.
This became problematic when trying to use the file either having invlade opcodes due to LLVM
or bad references to BPF/XDP maps.

To read an translate the object file to opcodes, use something like `llvm-objdump -Ss vpc_kern.o` to 
view the BPF opcodes and painfully convert them to the raw op codes found in the go code.
This can be sped up using a good IDE.

## BPF Opcode Basic structure
More info here: https://github.com/iovisor/bpf-docs/blob/master/eBPF.md and https://www.kernel.org/doc/Documentation/networking/filter.txt
```
msb                                                        lsb
+------------------------+----------------+----+----+--------+
|immediate               |offset          |src |dst |opcode  |
+------------------------+----------------+----+----+--------+
```
From least significant to most significant bit:

- 8 bit opcode
- 4 bit destination register (dst)
- 4 bit source register (src)
- 16 bit offset
- 32 bit immediate/constant (imm)

## Notes from LLVM
There are certain versions of llc/llvm in which Opcode 0x85 (call imm or "Function call") has the src & dst registers are set to R1. In most 
cases, this causes the program to be rejected and needs to be updated with a src & dst registers of R0.

Future results may differ!