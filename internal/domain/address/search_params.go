package address

type SearchParams struct {
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

func NewSearchParams(zipCode, state, city, neighborhood, street, num string, latitude, longitude, addInfo *string) *SearchParams {
	return &SearchParams{
		ZipCode:      zipCode,
		State:        state,
		City:         city,
		Neighborhood: neighborhood,
		Street:       street,
		Num:          num,
		Latitude:     latitude,
		Longitude:    longitude,
		AddInfo:      addInfo,
	}
}
