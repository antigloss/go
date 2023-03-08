/*
 *
 * Copyright (C) 2023 Antigloss Huang (https://github.com/antigloss) All rights reserved.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package conf

import (
	"crypto/rsa"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/mitchellh/mapstructure"
	"reflect"
)

func decodeHook(hook DecodeHook) mapstructure.DecodeHookFunc {
	return mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
		decoder(hook),
	)
}

func decoder(hook DecodeHook) func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		// Check if the data type matches the expected one
		if f.Kind() != reflect.String {
			return data, nil
		}

		switch t {
		case reflect.TypeOf(rsa.PublicKey{}):
			return jwt.ParseRSAPublicKeyFromPEM([]byte(data.(string)))
		case reflect.TypeOf(rsa.PrivateKey{}):
			return jwt.ParseRSAPrivateKeyFromPEM([]byte(data.(string)))
		}

		return hook(t, data.(string))
	}
}
