OBJS = vpc_kern.o

LLC ?= llc
CLANG ?= clang
INC_FLAGS = -nostdinc -isystem `$(CLANG) -print-file-name=include`
EXTRA_CFLAGS ?= -O2 -emit-llvm -g

# In case up-to-date headers are not installed locally in /usr/include,
# use source build.
linuxhdrs ?= /usr/src/linux-headers-`uname -r`

LINUXINCLUDE =  -I$(linuxhdrs)/arch/x86/include/uapi \
                -I$(linuxhdrs)/arch/x86/include/generated/uapi \
                -I$(linuxhdrs)/arch/x86/include/ \
                -I$(linuxhdrs)/include/generated/uapi \
                -I$(linuxhdrs)/include/uapi \
                -I$(linuxhdrs)/include

all: $(OBJS)

.PHONY: clean
clean:
	rm -f $(OBJS)

$(OBJS):  %.o:%.c
	$(CLANG) $(INC_FLAGS) \
		-D__KERNEL__ -D__EXPORTED_HEADERS__ \
		-I../include $(LINUXINCLUDE) \
		-Wno-unused-value -Wno-pointer-sign \
		-Wno-compare-distinct-pointer-types \
		-Wno-gnu-variable-sized-type-not-at-end \
		-Wno-address-of-packed-member -Wno-tautological-compare \
		-Wno-unknown-warning-option \
		$(EXTRA_CFLAGS) -c $< -o -| $(LLC) -march=bpf -filetype=obj -o $@