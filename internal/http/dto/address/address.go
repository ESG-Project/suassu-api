package address

type AddressOut struct {
	ID           string  `json:"id"`
	State        string  `json:"state"`
	ZipCode      string  `json:"zipCode"`
	City         string  `json:"city"`
	Neighborhood string  `json:"neighborhood"`
	Street       string  `json:"street"`
	Num          string  `json:"num"`
	Latitude     *string `json:"latitude,omitempty"`
	Longitude    *string `json:"longitude,omitempty"`
	AddInfo      *string `json:"addInfo,omitempty"`
}
