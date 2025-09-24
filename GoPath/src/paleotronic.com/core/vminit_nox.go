// +build nox

package core

import (
	"paleotronic.com/core/settings"
)

func (vm *VM) PreInitOverrides() {
	settings.UseDHGRForHGR[vm.Index] = true
	settings.UseVerticalBlend[vm.Index] = true
	_, _ = vm.ExecuteRequest("vm.gfx.setrendermode", "HGR", settings.VM_FLAT)
	_, _ = vm.ExecuteRequest("vm.gfx.setrendermode", "DHGR", settings.VM_FLAT)
}
