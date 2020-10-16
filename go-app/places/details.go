package places

// PlaceDetails is a response of details
type PlaceDetails struct {
	HTMLAttributions []interface{} `json:"html_attributions"`
	Result           Details       `json:"result"`
	Status           string        `json:"status"`
}

// Details is a part of format of API response
type Details struct {
	AddressComponents []struct {
		LongName  string   `json:"long_name"`
		ShortName string   `json:"short_name"`
		Types     []string `json:"types"`
	} `json:"address_components"`
	AdrAddress       string `json:"adr_address"`
	BusinessStatus   string `json:"business_status"`
	FormattedAddress string `json:"formatted_address"`
	Geometry         struct {
		Location LatLng `json:"location"`
		Viewport struct {
			Northeast LatLng `json:"northeast"`
			Southwest LatLng `json:"southwest"`
		} `json:"viewport"`
	} `json:"geometry"`
	Icon         string `json:"icon"`
	Name         string `json:"name"`
	OpeningHours struct {
		OpenNow bool `json:"open_now"`
		Periods []struct {
			Close struct {
				Day  int    `json:"day"`
				Time string `json:"time"`
			} `json:"close"`
			Open struct {
				Day  int    `json:"day"`
				Time string `json:"time"`
			} `json:"open"`
		} `json:"periods"`
		WeekdayText []string `json:"weekday_text"`
	} `json:"opening_hours"`
	Photos []*struct {
		Height           int      `json:"height"`
		HTMLAttributions []string `json:"html_attributions"`
		PhotoReference   string   `json:"photo_reference"`
		Width            int      `json:"width"`
	} `json:"photos"`
	PlaceID  string `json:"place_id"`
	PlusCode struct {
		CompoundCode string `json:"compound_code"`
		GlobalCode   string `json:"global_code"`
	} `json:"plus_code"`
	Rating    float64 `json:"rating"`
	Reference string  `json:"reference"`
	Reviews   []*struct {
		AuthorName              string `json:"author_name"`
		AuthorURL               string `json:"author_url"`
		ProfilePhotoURL         string `json:"profile_photo_url"`
		Rating                  int    `json:"rating"`
		RelativeTimeDescription string `json:"relative_time_description"`
		Text                    string `json:"text"`
		Time                    int    `json:"time"`
	} `json:"reviews"`
	Types            []string `json:"types"`
	URL              string   `json:"url"`
	UserRatingsTotal int      `json:"user_ratings_total"`
	UtcOffset        int      `json:"utc_offset"`
	Vicinity         string   `json:"vicinity"`
	Website          string   `json:"website"`
}

// MarshalPlace converts Details to Place
func (p *Details) MarshalPlace() Place {
	return Place{
		PlaceID:      p.PlaceID,
		Name:         p.Name,
		Rating:       p.Rating,
		GooglemapURI: p.URL,
	}
}
