package logging

import "github.com/rs/zerolog/log"

func CheckError(err error) {
	if err != nil {
		log.Error().Msg(err.Error())
	}
}
