package hyper

import (
	"encoding/xml"
	"fmt"
	"math"
	"time"

	hyperAPI "github.com/tcfw/vpc/pkg/api/v1/hyper"
	libvirt "libvirt.org/libvirt-go"
)

func memStats(d *libvirt.Domain) (*hyperAPI.MemStats, error) {
	mStats, err := d.MemoryStats(15, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get mem stats: %s", err)
	}

	stats := &hyperAPI.MemStats{}

	for _, stat := range mStats {
		switch libvirt.DomainMemoryStatTags(stat.Tag) {
		case libvirt.DOMAIN_MEMORY_STAT_SWAP_IN:
			stats.SwapIn = stat.Val
			break
		case libvirt.DOMAIN_MEMORY_STAT_SWAP_OUT:
			stats.SwapOut = stat.Val
			break
		case libvirt.DOMAIN_MEMORY_STAT_AVAILABLE:
			stats.Available = stat.Val
			break
		case libvirt.DOMAIN_MEMORY_STAT_UNUSED:
			stats.Unused = stat.Val
			break
		case libvirt.DOMAIN_MEMORY_STAT_MAJOR_FAULT:
			stats.MajorsFaults = stat.Val
			break
		case libvirt.DOMAIN_MEMORY_STAT_MINOR_FAULT:
			stats.MinorFaults = stat.Val
			break
		case libvirt.DOMAIN_MEMORY_STAT_USABLE:
			stats.Usable = stat.Val
			break
		case libvirt.DOMAIN_MEMORY_STAT_LAST_UPDATE:
			stats.LastUpdate = stat.Val
			break
		}
	}

	return stats, nil
}

func cpuStats(d *libvirt.Domain) ([]*hyperAPI.VCPUStats, *hyperAPI.VCPUStats, error) {
	stats := []*hyperAPI.VCPUStats{}
	total := &hyperAPI.VCPUStats{}

	tn := 1 * time.Second
	n, _ := d.GetMaxVcpus()

	p1, err := d.GetCPUStats(0, n, 0)
	if err != nil {
		return nil, nil, err
	}

	time.Sleep(tn)

	p2, err := d.GetCPUStats(0, n, 0)
	if err != nil {
		return nil, nil, err
	}

	for i := uint(0); i < n; i++ {
		useTimeDiff := math.Abs(float64(p2[i].CpuTime - p1[i].CpuTime))
		total.Time += uint64(useTimeDiff)

		vCPUStat := &hyperAPI.VCPUStats{
			Id:    uint32(i / 2),
			Time:  p1[i].CpuTime,
			Usage: float32(100*useTimeDiff) / float32(tn),
		}

		stats = append(stats, vCPUStat)
	}

	total.Usage = float32(100*total.Time) / float32(tn) / float32(n)

	return stats, total, nil
}

func netStats(d *libvirt.Domain) ([]*hyperAPI.NetStats, error) {
	stats := []*hyperAPI.NetStats{}

	ifaces, err := d.ListAllInterfaceAddresses(libvirt.DOMAIN_INTERFACE_ADDRESSES_SRC_ARP)
	if err != nil {
		return nil, err
	}

	for _, iface := range ifaces {
		ifaceStats, err := d.InterfaceStats(iface.Name)
		if err != nil {
			return nil, err
		}

		stats = append(stats, &hyperAPI.NetStats{
			Id:      iface.Name,
			RxBytes: uint64(ifaceStats.RxBytes),
			RxPkts:  uint64(ifaceStats.RxPackets),
			RxErrs:  uint64(ifaceStats.RxErrs),
			RxDrops: uint64(ifaceStats.RxDrop),
			TxBytes: uint64(ifaceStats.TxBytes),
			TxPkts:  uint64(ifaceStats.TxPackets),
			TxErrs:  uint64(ifaceStats.TxErrs),
			TxDrops: uint64(ifaceStats.TxDrop),
		})
	}

	return stats, nil
}

func diskStats(d *libvirt.Domain) ([]*hyperAPI.DiskStats, *hyperAPI.DiskStats, error) {
	stats := []*hyperAPI.DiskStats{}
	total := &hyperAPI.DiskStats{}

	r, err := d.GetXMLDesc(libvirt.DOMAIN_XML_MIGRATABLE)
	if err != nil {
		return nil, nil, err
	}

	desc := &DomainDesc{}

	xml.Unmarshal([]byte(r), desc)

	for _, disk := range desc.Devices.Disks {
		dStats, err := d.BlockStats(disk.Target.Dev)
		if err != nil {
			return nil, nil, err
		}

		stats = append(stats, &hyperAPI.DiskStats{
			Id:      disk.Target.Dev,
			RdReqs:  uint64(dStats.RdReq),
			RdBytes: uint64(dStats.RdBytes),
			RdTimes: uint64(dStats.RdTotalTimes),
			WrReqs:  uint64(dStats.WrReq),
			WrBytes: uint64(dStats.WrBytes),
			WrTimes: uint64(dStats.WrTotalTimes),
		})

		total.RdReqs += uint64(dStats.RdReq)
		total.RdBytes += uint64(dStats.RdBytes)
		total.RdTimes += uint64(dStats.RdTotalTimes)
		total.WrReqs += uint64(dStats.WrReq)
		total.WrBytes += uint64(dStats.WrBytes)
		total.WrTimes += uint64(dStats.WrTotalTimes)
	}

	return stats, total, nil
}
