package address

type Address struct {
	ID           string
	ZipCode      string
	State        string
	City         string
	Neighborhood string
	Street       string
	Num          string
	Latitude     *string
	Longitude    *string
	AddInfo      *string
}

func NewAddress(id, zipCode, state, city, neighborhood, street, num string) *Address {
	return &Address{
		ID:           id,
		ZipCode:      zipCode,
		State:        state,
		City:         city,
		Neighborhood: neighborhood,
		Street:       street,
		Num:          num,
	}
}

func (u *Address) SetLatitude(latitude *string) {
	u.Latitude = latitude
}

func (u *Address) SetLongitude(longitude *string) {
	u.Longitude = longitude
}

func (u *Address) SetAddInfo(addInfo *string) {
	u.AddInfo = addInfo
}
