package hyper

import (
	"fmt"

	hyperAPI "github.com/tcfw/vpc/pkg/api/v1/hyper"
	libvirt "libvirt.org/libvirt-go"
)

func setDesirePower(d *libvirt.Domain, desiredState hyperAPI.PowerState) (hyperAPI.PowerState, error) {
	current, _, err := d.GetState()
	if err != nil {
		return hyperAPI.PowerState_NONE, err
	}

	switch current {
	case libvirt.DOMAIN_CRASHED:
	case libvirt.DOMAIN_BLOCKED:
		switch desiredState {
		case hyperAPI.PowerState_RUNNING:
			return appliedState(hyperAPI.PowerState_CRASHED, d.Reboot(libvirt.DOMAIN_REBOOT_DEFAULT))
		case hyperAPI.PowerState_SHUTDOWN:
			return appliedState(hyperAPI.PowerState_SHUTOFF, d.Destroy())
		}
	case libvirt.DOMAIN_RUNNING:
		switch desiredState {
		case hyperAPI.PowerState_SHUTDOWN:
			return appliedState(hyperAPI.PowerState_SHUTDOWN, d.ShutdownFlags(libvirt.DOMAIN_SHUTDOWN_ACPI_POWER_BTN))
		case hyperAPI.PowerState_SHUTOFF:
			return appliedState(hyperAPI.PowerState_SHUTOFF, d.DestroyFlags(libvirt.DOMAIN_DESTROY_GRACEFUL))
		}
	case libvirt.DOMAIN_PAUSED:
		switch desiredState {
		case hyperAPI.PowerState_SHUTDOWN:
			return appliedState(hyperAPI.PowerState_SHUTOFF, d.Shutdown())
		case hyperAPI.PowerState_RUNNING:
			return appliedState(hyperAPI.PowerState_RUNNING, d.Resume())
		}
	case libvirt.DOMAIN_SHUTDOWN:
		switch desiredState {
		case hyperAPI.PowerState_SHUTOFF:
			return appliedState(hyperAPI.PowerState_SHUTOFF, d.DestroyFlags(libvirt.DOMAIN_DESTROY_DEFAULT))
		}
	case libvirt.DOMAIN_SHUTOFF:
		switch desiredState {
		case hyperAPI.PowerState_RUNNING:
			return appliedState(hyperAPI.PowerState_RUNNING, d.Create())
		}
	}

	return libVirtToAPI(current), nil
}

func appliedState(endCurrent hyperAPI.PowerState, err error) (hyperAPI.PowerState, error) {
	if err != nil {
		return hyperAPI.PowerState_NONE, fmt.Errorf("failed to apply state: %s", err)
	}

	return endCurrent, nil
}

func libVirtToAPI(state libvirt.DomainState) hyperAPI.PowerState {
	switch state {
	case libvirt.DOMAIN_CRASHED:
	case libvirt.DOMAIN_BLOCKED:
		return hyperAPI.PowerState_CRASHED
	case libvirt.DOMAIN_RUNNING:
		return hyperAPI.PowerState_RUNNING
	case libvirt.DOMAIN_SHUTOFF:
		return hyperAPI.PowerState_SHUTOFF
	case libvirt.DOMAIN_SHUTDOWN:
		return hyperAPI.PowerState_SHUTDOWN
	case libvirt.DOMAIN_PAUSED:
	case libvirt.DOMAIN_PMSUSPENDED:
		return hyperAPI.PowerState_MIGRATING
	case libvirt.DOMAIN_NOSTATE:
	default:
		//
	}

	return hyperAPI.PowerState_NONE
}
