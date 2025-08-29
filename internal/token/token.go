package token

import "errors"

type Audiance int

const (
	AudianceAdmin Audiance = iota + 1
	AudianceUser
)

func (a Audiance) String() string {
	switch a {
	case AudianceAdmin:
		return "admin"
	case AudianceUser:
		return "user"
	default:
		return ""
	}
}

func (a Audiance) Int() int {
	return int(a)
}

func ParseAudiance(aud string) (Audiance, error) {
	switch aud {
	case "admin":
		return AudianceAdmin, nil
	case "user":
		return AudianceUser, nil
	default:
		return -1, errors.New("invalid audiance")
	}
}
