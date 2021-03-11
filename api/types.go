package api

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
)
