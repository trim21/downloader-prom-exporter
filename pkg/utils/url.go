package utils

import (
	"net/url"
	"strconv"

	"github.com/sirupsen/logrus"
)

func GetPort(u *url.URL) uint16 {
	if r := u.Port(); r != "" {
		v, err := strconv.Atoi(r)
		if err != nil {
			logrus.Fatalln(v)
		}

		return uint16(v)
	}

	if u.Scheme == "https" {
		return 443
	}

	return 80
}

func GetUserPass(u *url.Userinfo) (username string, password string) {
	if u == nil {
		return
	}

	username = u.Username()
	password, _ = u.Password()

	return
}
