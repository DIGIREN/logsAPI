package models

type LogObject struct {
	UserID int     `json:"user_id"`
	Total  float64 `json:"total"`
	Title  string  `json:"title"`
	Meta   struct {
		Logins []struct {
			Time string `json:"time"`
			IP   string `json:"ip"`
		} `json:"logins"`
		PhoneNumbers struct {
			Home   string `json:"home"`
			Mobile string `json:"mobile"`
		} `json:"phone_numbers"`
	} `json:"meta"`
	Completed bool `json:"completed"`
}
