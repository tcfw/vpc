package hyper

import (
	"encoding/xml"
	"testing"

	"aqwari.net/xml/xmltree"
	"github.com/stretchr/testify/assert"
)

func TestMarshalUnmarshal(t *testing.T) {
	obj := &DomainDesc{}
	sample := []byte(`<domain type="kvm">
  <name>vs-1</name>
  <uuid>f5a921ce-eedd-4d99-8a2a-2a98f4470ca8</uuid>
  <metadata>
    <libosinfo:libosinfo xmlns:libosinfo="http://libosinfo.org/xmlns/libvirt/domain/1.0">
      <libosinfo:os id="http://ubuntu.com/ubuntu/18.04"/>
    </libosinfo:libosinfo>
  </metadata>
  <memory unit="KiB">2097152</memory>
  <currentMemory unit="KiB">2097152</currentMemory>
  <vcpu placement="static">2</vcpu>
  <os>
    <type arch="x86_64" machine="pc-q35-3.1">hvm</type>
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
  <on_crash>destroy</on_crash>
  <pm>
    <suspend-to-mem enabled="no"/>
    <suspend-to-disk enabled="no"/>
  </pm>
  <devices>
    <emulator>/usr/bin/qemu-system-x86_64</emulator>
    <disk type="file" device="disk">
      <driver name="qemu" type="qcow2"/>
      <source file="/var/lib/libvirt/images/vs-1.qcow2"/>
      <target dev="vda" bus="virtio"/>
      <address type="pci" domain="0x0000" bus="0x03" slot="0x00" function="0x0"/>
    </disk>
    <controller type="sata" index="0">
      <address type="pci" domain="0x0000" bus="0x00" slot="0x1f" function="0x2"/>
    </controller>
    <controller type="pci" index="0" model="pcie-root"/>
    <controller type="pci" index="1" model="pcie-root-port">
      <model name="pcie-root-port"/>
      <target chassis="1" port="0x10"/>
      <address type="pci" domain="0x0000" bus="0x00" slot="0x02" function="0x0" multifunction="on"/>
    </controller>
    <controller type="virtio-serial" index="0">
      <address type="pci" domain="0x0000" bus="0x02" slot="0x00" function="0x0"/>
    </controller>
    <interface type="network">
      <mac address="52:54:00:d6:58:37"/>
      <source network="default"/>
      <model type="e1000e"/>
      <address type="pci" domain="0x0000" bus="0x06" slot="0x00" function="0x0"/>
    </interface>
    <serial type="pty">
      <target type="isa-serial" port="0">
        <model name="isa-serial"/>
      </target>
    </serial>
    <console type="pty">
      <target type="serial" port="0"/>
    </console>
    <channel type="unix">
      <target type="virtio" name="org.qemu.guest_agent.0"/>
      <address type="virtio-serial" controller="0" bus="0" port="1"/>
    </channel>
    <channel type="spicevmc">
      <target type="virtio" name="com.redhat.spice.0"/>
      <address type="virtio-serial" controller="0" bus="0" port="2"/>
    </channel>
    <input type="tablet" bus="usb">
      <address type="usb" bus="0" port="1"/>
    </input>
    <input type="mouse" bus="ps2"/>
    <input type="keyboard" bus="ps2"/>
    <graphics type="spice" autoport="yes">
      <listen type="address"/>
      <image compression="off"/>
    </graphics>
    <video>
      <model type="qxl" ram="65536" vram="65536" vgamem="16384" heads="1" primary="yes"/>
      <address type="pci" domain="0x0000" bus="0x00" slot="0x01" function="0x0"/>
    </video>
    <memballoon model="virtio">
      <address type="pci" domain="0x0000" bus="0x04" slot="0x00" function="0x0"/>
    </memballoon>
    <rng model="virtio">
      <backend model="random">/dev/urandom</backend>
      <address type="pci" domain="0x0000" bus="0x05" slot="0x00" function="0x0"/>
    </rng>
  </devices>
</domain>
`)

	err := xml.Unmarshal(sample, obj)
	if err != nil {
		t.Fatal(err)
	}

	b, err := xml.Marshal(obj)
	if err != nil {
		t.Fatal(err)
	}

	xmlEla, err := xmltree.Parse(sample)
	if err != nil {
		t.Fatal(err)
	}
	xmlElb, err := xmltree.Parse(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.True(t, xmltree.Equal(xmlEla, xmlElb))
}
