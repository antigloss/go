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
	"reflect"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/mitchellh/mapstructure"
	"google.golang.org/protobuf/types/known/durationpb"
)

func decodeHook(hooks []DecodeHook) mapstructure.DecodeHookFunc {
	return mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
		decoder(hooks),
	)
}

func decoder(hooks []DecodeHook) func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	hm := map[reflect.Type]func(string) (interface{}, error){}
	hm[reflect.TypeOf(rsa.PublicKey{})] = func(data string) (interface{}, error) {
		return jwt.ParseRSAPublicKeyFromPEM([]byte(data))
	}
	hm[reflect.TypeOf(rsa.PrivateKey{})] = func(data string) (interface{}, error) {
		return jwt.ParseRSAPrivateKeyFromPEM([]byte(data))
	}
	hm[reflect.TypeOf(durationpb.Duration{})] = func(data string) (interface{}, error) {
		t, e := time.ParseDuration(data)
		if e != nil {
			return nil, e
		}
		return durationpb.New(t), nil
	}
	for _, hook := range hooks {
		hm[hook.Type] = hook.Decode
	}

	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		// Check if the data type matches the expected one
		if f.Kind() != reflect.String {
			return data, nil
		}

		if h, ok := hm[t]; ok {
			return h(data.(string))
		}

		return data, nil
	}
}
