package hyper

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	hyperAPI "github.com/tcfw/vpc/pkg/api/v1/hyper"
)

func TestDefineVM(t *testing.T) {
	vm := &hyperAPI.VM{
		Id: "test-1",
		TemplateId: `<domain type="kvm">
	<memory unit="KiB">4194304</memory>
	<currentMemory unit="KiB">4194304</currentMemory>
	<vcpu placement="static">2</vcpu>
	<os>
		<type arch="x86_64" machine="pc-q35-4.2">hvm</type>
		<boot dev="hd"/>
	</os>
	<features>
		<acpi/>
		<apic/>
		<vmport state="off"/>
	</features>
	<cpu mode="host-model" check="partial"/>
	<clock offset="utc">
		<timer name="rtc" tickpolicy="catchup"/>
		<timer name="pit" tickpolicy="delay"/>
		<timer name="hpet" present="no"/>
	</clock>
	<on_poweroff>destroy</on_poweroff>
	<on_reboot>restart</on_reboot>
	<on_crash>restart</on_crash>
	<pm>
		<suspend-to-mem enabled="no"/>
		<suspend-to-disk enabled="no"/>
	</pm>
	<devices>
		<emulator>/usr/bin/qemu-system-x86_64</emulator>
		<controller type="pci" index="0" model="pcie-root"/>
		<controller type="pci" index="1" model="pcie-root-port">
			<model name="pcie-root-port"/>
			<target chassis="1" port="0x10"/>
			<address type="pci" domain="0x0000" bus="0x00" slot="0x02" function="0x0" multifunction="on"/>
		</controller>
		<controller type="pci" index="2" model="pcie-root-port">
			<model name="pcie-root-port"/>
			<target chassis="2" port="0x11"/>
			<address type="pci" domain="0x0000" bus="0x00" slot="0x02" function="0x1"/>
		</controller>
		<controller type="pci" index="3" model="pcie-to-pci-bridge">
			<model name="pcie-pci-bridge"/>
			<address type="pci" domain="0x0000" bus="0x01" slot="0x00" function="0x0"/>
		</controller>
		<controller type="pci" index="4" model="pcie-root-port">
			<model name="pcie-root-port"/>
			<target chassis="4" port="0x12"/>
			<address type="pci" domain="0x0000" bus="0x00" slot="0x02" function="0x2"/>
		</controller>
		<controller type="pci" index="5" model="pcie-root-port">
			<model name="pcie-root-port"/>
			<target chassis="5" port="0x13"/>
			<address type="pci" domain="0x0000" bus="0x00" slot="0x02" function="0x3"/>
		</controller>
		<controller type="pci" index="6" model="pcie-root-port">
			<model name="pcie-root-port"/>
			<target chassis="6" port="0x14"/>
			<address type="pci" domain="0x0000" bus="0x00" slot="0x02" function="0x4"/>
		</controller>
		<controller type="pci" index="7" model="pcie-root-port">
			<model name="pcie-root-port"/>
			<target chassis="7" port="0x15"/>
			<address type="pci" domain="0x0000" bus="0x00" slot="0x02" function="0x5"/>
		</controller>
		<controller type="pci" index="8" model="pcie-root-port">
			<model name="pcie-root-port"/>
			<target chassis="8" port="0x16"/>
			<address type="pci" domain="0x0000" bus="0x00" slot="0x02" function="0x6"/>
		</controller>
		<controller type="pci" index="9" model="pcie-root-port">
			<model name="pcie-root-port"/>
			<target chassis="9" port="0x8"/>
			<address type="pci" domain="0x0000" bus="0x00" slot="0x01" function="0x0"/>
		</controller>
		<controller type="virtio-serial" index="0">
			<address type="pci" domain="0x0000" bus="0x04" slot="0x00" function="0x0"/>
		</controller>
		<serial type="pty">
			<target type="isa-serial" port="0">
				<model name="isa-serial"/>
			</target>
		</serial>
		<console type="pty">
			<target type="serial" port="0"/>
		</console>
		<memballoon model="virtio">
			<address type="pci" domain="0x0000" bus="0x06" slot="0x00" function="0x0"/>
		</memballoon>
		<rng model="virtio">
			<backend model="random">/dev/urandom</backend>
			<address type="pci" domain="0x0000" bus="0x07" slot="0x00" function="0x0"/>
		</rng>
	</devices>
</domain>`,
	}

	c, err := testLibVirtConn()
	if err != nil {
		log.Fatalf("failed: %s", err)
	}

	defer c.Close()

	err = DefineVM(c, vm)
	if err != nil {
		log.Fatal(err)
	}

	assert.NotEmpty(t, vm.HyperId)
}
