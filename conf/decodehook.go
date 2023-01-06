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
