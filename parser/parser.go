package parser

import (
	"fmt"
	"reflect"

	"github.com/glycerine/zygomys/zygo"
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

func SexpToGo(sexp zygo.Sexp) (interface{}, error) {
	switch v := sexp.(type) {
	case *zygo.SexpStr:
		return v.S, nil
	case *zygo.SexpInt:
		return v.Val, nil
	case *zygo.SexpFloat:
		return v.Val, nil
	case *zygo.SexpBool:
		return v.Val, nil
	case *zygo.SexpChar:
		return v.Val, nil
	case *zygo.SexpRaw:
		return v.Val, nil
	case *zygo.SexpArray:
		items := make([]interface{}, len(v.Val))
		for i, item := range v.Val {
			val, err := SexpToGo(item)
			if err != nil {
				return nil, err
			}
			items[i] = val
		}
		return items, nil
	case *zygo.SexpHash:
		result := make(map[string]interface{})
		for _, pairList := range v.Map {
			for _, pair := range pairList {
				kstr := pair.Head.(*zygo.SexpSymbol)
				val, err := SexpToGo(pair.Tail)
				if err != nil {
					return nil, err
				}
				result[kstr.Name()] = val
			}
		}
		return result, nil
	default:
		if sexp == zygo.SexpNull {
			return nil, nil
		}
		return nil, fmt.Errorf("unsupported S-expression type: %T", sexp)
	}
}
