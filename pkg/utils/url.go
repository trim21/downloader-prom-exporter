package utils

import (
	"net/url"
	"strconv"

	"github.com/rs/zerolog/log"
)

func GetPort(u *url.URL) uint16 {
	if r := u.Port(); r != "" {
		v, err := strconv.Atoi(r)
		if err != nil {
			log.Fatal().Int("port", v).Err(err).Msg("failed to parse transmission port")
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
