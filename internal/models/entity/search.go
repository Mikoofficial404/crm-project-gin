package entity

type GlobalSearchResponse struct {
	Leads    []Lead    `json:"leads"`
	Deals    []Deal    `json:"deals"`
	Users    []User    `json:"users"`
	Contacts []Contact `json:"contacts"`
}
