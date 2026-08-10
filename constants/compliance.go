package constants

const (
	ComplianceStatusPending  = "pending"
	ComplianceStatusApproved = "approved"
	ComplianceStatusRejected = "rejected"
)

type SIPPerson struct {
	Name        string `json:"name"`
	DateOfBirth string `json:"date_of_birth"`
	Country     string `json:"country"`
}

// SIPList contains fictional test records used for internal compliance screening.
// A merchant is flagged only when name, date of birth, and country all match.
var SIPList = []SIPPerson{
	{
		Name:        "John Smith",
		DateOfBirth: "1972-09-10",
		Country:     "UK",
	},
	{
		Name:        "Jane Doe",
		DateOfBirth: "1985-03-15",
		Country:     "US",
	},
	{
		Name:        "Robert Example",
		DateOfBirth: "1968-11-22",
		Country:     "CA",
	},
	{
		Name:        "Michael Test",
		DateOfBirth: "1975-07-18",
		Country:     "AU",
	},
	{
		Name:        "Sarah Sample",
		DateOfBirth: "1982-12-03",
		Country:     "DE",
	},
}

// RestrictedCountries represents SamPay's internal V1 restricted-country policy.
// This is not intended to be a complete or authoritative sanctions list.
var RestrictedCountries = []string{
	"North Korea",
	"Iran",
	"Syria",
	"Sudan",
	"Cuba",
}
