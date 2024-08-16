package environment

import "sync"

var (
	buildingStatus = false
	patchStatus    = false
	buildMutex     = &sync.RWMutex{}
)

func SetBuildingStatus(status bool) {
	buildMutex.Lock()
	defer buildMutex.Unlock()
	buildingStatus = status
}

func IsBuilding() bool {
	buildMutex.RLock()
	defer buildMutex.RUnlock()
	return buildingStatus
}

func SetPatchStatus(status bool) {
	buildMutex.Lock()
	defer buildMutex.Unlock()
	patchStatus = status
}

func IsPatch() bool {
	buildMutex.RLock()
	defer buildMutex.RUnlock()
	return patchStatus
}
