package interfaces

import (
	"paleotronic.com/core/types"
)

type VM interface {
	GetMemory(addr int) uint64
	SetMemory(addr int, value uint64)
	GetHUDLayerByID(name string) (*types.LayerSpecMapped, bool)
	GetGFXLayerByID(name string) (*types.LayerSpecMapped, bool)
	GetHUDLayerSet() []*types.LayerSpecMapped
	GetGFXLayerSet() []*types.LayerSpecMapped
	DisableGFXLayers() map[string]bool
	EnableGFXLayers(enabled map[string]bool)
	ExecutePendingTasks()
	ExecuteRequest(action string, args ...interface{}) (interface{}, error)
	Logf(pattern string, args ...interface{})
	IsDying() bool
	GetLayers() ([]*types.LayerSpecMapped, []*types.LayerSpecMapped)
	Teardown()
}
