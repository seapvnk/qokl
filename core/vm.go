package core

import (
	"fmt"
	"os"
	"reflect"

	"github.com/glycerine/zygomys/zygo"
)

type ZygResult struct {
	Value zygo.Sexp
	Error error
}

func ExecuteScript(path string, input map[string]any) (*ZygResult, error) {
	z := zygo.NewZlisp()

	if input != nil {
		for k, v := range input {
			z.AddGlobal(k, toSexp(z, v))
		}
	}

	code, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read file %s: %w", path, err)
	}

	err = z.LoadString(string(code))
	if err != nil {
		return nil, fmt.Errorf("error executing %s: %w", path, err)
	}

	out, err := z.Run()
	return &ZygResult{Value: out, Error: err}, nil
}

func toSexp(env *zygo.Zlisp, val interface{}) zygo.Sexp {
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
				arr[i] = toSexp(env, rv.Index(i).Interface())
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
				valSexp := toSexp(env, valInterface)
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
