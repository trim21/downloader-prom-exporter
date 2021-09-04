package utils

import (
	"os"
)

func Exist(name string) (bool, error) {
	_, err := os.Stat(name)
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
