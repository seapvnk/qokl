package parser

import (
	"fmt"
	"reflect"

	"github.com/glycerine/zygomys/v9/zygo"
)

func ToSexp(env *zygo.Zlisp, val interface{}) zygo.Sexp {
	switch v := val.(type) {
	case string:
		return &zygo.SexpStr{S: v}
	case int:
		return &zygo.SexpInt{Val: int64(v)}
	case int64:
		return &zygo.SexpInt{Val: v}
	case float64:
		return &zygo.SexpFloat{Val: v}
	case float32:
		return &zygo.SexpFloat{Val: float64(v)}
	case bool:
		return &zygo.SexpBool{Val: v}
	case rune:
		return &zygo.SexpChar{Val: v}
	case []byte:
		return &zygo.SexpRaw{Val: v}
	default:
		rv := reflect.ValueOf(val)
		if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
			arr := make([]zygo.Sexp, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				arr[i] = ToSexp(env, rv.Index(i).Interface())
			}
			return &zygo.SexpArray{Val: arr}
		}

		if rv.Kind() == reflect.Map {
			hash := zygo.SexpHash{
				Map: make(map[int][]*zygo.SexpPair),
			}
			for _, key := range rv.MapKeys() {
				kInterface := key.Interface()
				k, ok := kInterface.(string)
				if !ok {
					panic(fmt.Sprintf("map key is not string: %T", kInterface))
				}
				valInterface := rv.MapIndex(key).Interface()
				keySexp := env.MakeSymbol(k)
				valSexp := ToSexp(env, valInterface)
				err := hash.HashSet(keySexp, valSexp)
				if err != nil {
					panic(fmt.Sprintf("error setting hash key %q: %v", k, err))
				}
			}
			return &hash
		}
		// fallback: string
		return &zygo.SexpStr{S: fmt.Sprintf("%v", val)}
	}
}
