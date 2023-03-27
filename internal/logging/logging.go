package logging

import "github.com/rs/zerolog/log"

func CheckError(err error) error {
	if err != nil {
		log.Error().Msg(err.Error())
	}
	return err
}
