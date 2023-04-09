// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package logging

import "github.com/rs/zerolog/log"

func CheckError(err error) error {
	if err != nil {
		log.Error().Msg(err.Error())
	}
	return err
}
