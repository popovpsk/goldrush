package types

//go:generate easyjson

//easyjson:json
type (
	Area struct {
		PosX  int32 `json:"posX"`
		PosY  int32 `json:"posY"`
		SizeX int32 `json:"sizeX"`
		SizeY int32 `json:"sizeY"`
	}

	ExploredArea struct {
		Area   Area  `json:"area"`
		Amount int32 `json:"amount"`
	}

	License struct {
		ID         int32 `json:"id"`
		DigAllowed int32 `json:"digAllowed"`
		DigUsed    int32 `json:"digUsed"`
	}

	PostLicenseRequest []uint32

	DigRequest struct {
		LicenseID int32 `json:"licenseID"`
		PosX      int32 `json:"posX"`
		PosY      int32 `json:"posY"`
		Depth     int32 `json:"depth"`
	}

	Treasures []string

	Payment []uint32
)

func (v *Area) Size() int32 {
	return v.SizeX * v.SizeY
}
