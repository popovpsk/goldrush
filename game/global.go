package game

import (
	"goldrush/types"
	"sync"
)

var areaPool sync.Pool

func AcquireArea() *types.Area {
	a := areaPool.Get()
	if a == nil {
		return &types.Area{}
	}
	return a.(*types.Area)
}

func ReleaseArea(area *types.Area) {
	area.PosX, area.PosY, area.SizeX, area.SizeY = 0, 0, 0, 0
	areaPool.Put(area)
}

var exploredAreaPool sync.Pool

func AcquireExploredArea() *types.ExploredArea {
	a := exploredAreaPool.Get()
	if a == nil {
		return &types.ExploredArea{}
	}
	return a.(*types.ExploredArea)
}

func ReleaseExploredArea(a *types.ExploredArea) {
	a.Amount, a.Area.PosX, a.Area.PosY, a.Area.SizeX, a.Area.SizeY = 0, 0, 0, 0, 0
	exploredAreaPool.Put(a)
}
