package api

import (
	"fmt"
	"sort"
)

//go:generate easyjson

//easyjson:json
type (
	Area struct {
		PosX  int `json:"posX"`
		PosY  int `json:"posY"`
		SizeX int `json:"sizeX"`
		SizeY int `json:"sizeY"`
	}

	ExploreResponse struct {
		Area   Area `json:"area"`
		Amount int  `json:"amount"`
	}

	LicensesResponse []License

	License struct {
		ID         int `json:"id"`
		DigAllowed int `json:"digAllowed"`
		DigUsed    int `json:"digUsed"`
	}

	PostLicenseRequest []uint32

	DigRequest struct {
		LicenseID int   `json:"licenseID"`
		PosX      int32 `json:"posX"`
		PosY      int32 `json:"posY"`
		Depth     int   `json:"depth"`
	}

	Treasures []string

	Payment []uint32

	BalanceResponse struct {
		Balance uint32
		Wallet  []uint32
	}

	Point struct {
		X, Y, Amount int32
	}

	Output []PointInfo

	PointInfo struct {
		X, Y  int32
		Depth int32
		Money int32
	}
)

func (v *Area) Size() int {
	return v.SizeX * v.SizeY
}

func (o *Output) Result() string {
	result := ""
	sort.Slice(*o, func(i, j int) bool {
		return (*o)[i].Money > (*o)[j].Money
	})
	for _, v := range *o {
		result += fmt.Sprintf("%v:%v:%v|", v.X, v.Y, v.Money)
	}
	return result
}

func (p *PointInfo) GetKey() string {
	return fmt.Sprintf("%v:%v", p.X, p.Y)
}
