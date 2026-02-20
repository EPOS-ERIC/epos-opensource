package common

type URLs struct {
	GUIURL        string
	APIURL        string
	BackofficeURL *string
}

func (u *URLs) String() string {
	if u == nil {
		return ""
	}

	if u.BackofficeURL == nil || *u.BackofficeURL == "" {
		return "gui=" + u.GUIURL + ", api=" + u.APIURL
	}

	return "gui=" + u.GUIURL + ", api=" + u.APIURL + ", backoffice=" + *u.BackofficeURL
}
