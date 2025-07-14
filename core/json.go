package core

import (
	"fmt"

	"github.com/glycerine/zygomys/zygo"
)

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

