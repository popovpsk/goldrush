package api

import (
	"goldrush/types"
	"strconv"
)

type JsonS struct {
	digPool chan *[]byte
	expPool chan *[]byte
}

func NewJsonS() *JsonS {
	digPool := make(chan *[]byte, parallelRequests*2)
	for i := 0; i < parallelRequests*2; i++ {
		b := make([]byte, 0, 64)
		b = append(b, "{\"licenseID\":"...)
		digPool <- &b
	}
	expPool := make(chan *[]byte, parallelRequests*2)
	for i := 0; i < parallelRequests*2; i++ {
		b := make([]byte, 0, 64)
		b = append(b, "{\"posX\":"...)
		expPool <- &b
	}

	return &JsonS{
		digPool: digPool,
		expPool: expPool,
	}
}

func (j *JsonS) GetDigRequest(dr *types.DigRequest) *[]byte {
	b := <-j.digPool
	*b = (*b)[:13]
	*b = append(*b, strconv.Itoa(int(dr.LicenseID))...)
	*b = append(*b, ",\"posX\":"...)
	*b = append(*b, strconv.Itoa(int(dr.PosX))...)
	*b = append(*b, ",\"posY\":"...)
	*b = append(*b, strconv.Itoa(int(dr.PosY))...)
	*b = append(*b, ",\"depth\":"...)
	*b = append(*b, strconv.Itoa(int(dr.Depth))...)
	*b = append(*b, "}"...)
	return b
}

func (j *JsonS) GetExploreRequest(dr *types.Area) *[]byte {
	b := <-j.expPool
	*b = (*b)[:8]
	*b = append(*b, strconv.Itoa(int(dr.PosX))...)
	*b = append(*b, ",\"posY\":"...)
	*b = append(*b, strconv.Itoa(int(dr.PosY))...)
	*b = append(*b, ",\"sizeX\":"...)
	*b = append(*b, strconv.Itoa(int(dr.SizeX))...)
	*b = append(*b, ",\"sizeY\":"...)
	*b = append(*b, strconv.Itoa(int(dr.SizeY))...)
	*b = append(*b, "}"...)
	return b
}

func (j *JsonS) ReleaseDig(b *[]byte) {
	j.digPool <- b
}

func (j *JsonS) ReleaseExp(b *[]byte) {
	j.expPool <- b
}
