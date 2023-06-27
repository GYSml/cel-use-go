package cel

import (
	"errors"
	"fmt"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/common/types/traits"
	"github.com/google/cel-go/ext"
	"reflect"
	"sync"
)

// CelConf 结构体配置
type CelConf struct {
	sync.Mutex
	celEnv *cel.Env  // cel变量处理
	celFields map[string]*cel.Type  // 1、用于普通变量注册  2、结构体变量注册
	celFunc  []cel.EnvOption // TODO:1、函数方法注册
}

var celConf = CelConf{
	celFields: map[string]*cel.Type{
		"M":cel.IntType,
		"N":cel.IntType,
	},
	celFunc: []cel.EnvOption{
		// 求和
		cel.Function("sum",
			cel.MemberOverload("list_int_sum", []*cel.Type{cel.ListType(cel.IntType)},
				cel.IntType, cel.UnaryBinding(func(value ref.Val) ref.Val {
					list, ok := value.(traits.Lister)
					if !ok {
						return types.IntNegOne
					}
					size, ok := list.Size().Value().(int64)
					if !ok {
						return types.IntNegOne
					}
					res := 0
					for i := 0; i < int(size); i++ {
						v, ok := list.Get(types.Int(i)).Value().(int64)
						if !ok {
							v = 0
						}
						res += int(v)
					}
					return types.Int(res)
				}))),
	},
}

func InitCleConf()  error{
	var err error
	celConf.Lock()
	defer celConf.Unlock()
	opts := make([]cel.EnvOption, 0, len(celConf.celFunc) + len(celConf.celFields))
	for key, val := range celConf.celFields {
		opts = append(opts, cel.Variable(key, val))
	}
	opts = append(opts, celConf.celFunc...)
	celConf.celEnv, err = cel.NewEnv(opts...)
	if err != nil {
		return err
	}
	return nil
}


func TypeToCELType(typ reflect.Type) (*cel.Type, error) {
	switch typ.Kind() {
	case reflect.Ptr:
		return TypeToCELType(typ.Elem())
	case reflect.Bool:
		return cel.BoolType, nil
	case reflect.Float64, reflect.Float32:
		return cel.DoubleType, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return cel.IntType, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return cel.UintType, nil
	case reflect.String:
		return cel.StringType, nil
	case reflect.Slice:
		if typ.Elem().Kind() == reflect.Uint8 {
			return cel.BytesType, nil
		}
		fallthrough
	case reflect.Array:
		elem, err := TypeToCELType(typ.Elem())
		if err != nil {
			return nil, err
		}
		return cel.ListType(elem), nil
	case reflect.Map:
		key, err := TypeToCELType(typ.Key())
		if err != nil {
			return nil, err
		}
		val, err := TypeToCELType(typ.Elem())
		if err != nil {
			return nil, err
		}
		return cel.MapType(key, val), nil
	case reflect.Struct:
		return cel.ObjectType(typ.String()), nil
	case reflect.Interface:
		return cel.DynType, nil
	}

	return nil, fmt.Errorf("unsupported type conversion kind %s", typ.Kind())
}


// calculate 表达式计算
func calculate(str string, varsMap map[string]interface{}) (ref.Val, error) {
	//  未知字段支持
	var err error
	celConf.Lock()
	defaultEnv, err := celConf.celEnv.Extend()
	if err != nil {
		return nil, err
	}
	for key, v := range varsMap {
		val := v
		if _, ok := celConf.celFields[key]; !ok {
			tp := reflect.TypeOf(val)
			var opt []cel.EnvOption
			if tp.Kind() == reflect.Struct {
				opt = append(opt, ext.NativeTypes(reflect.TypeOf(val)), cel.Variable(key, cel.ObjectType(tp.String())))

			} else {
				ctp, err := TypeToCELType(tp)
				if err != nil {
					continue
				}
				opt = append(opt, cel.Variable(key, ctp))
			}
			extEnv, err := defaultEnv.Extend(opt...)
			if err != nil {
				continue
			} else {
				defaultEnv = extEnv
			}
		}
	}
	celConf.Unlock()

	if defaultEnv == nil {
		return nil, errors.New("env is nil use InitEnv for init")
	}
	ast, issues := defaultEnv.Compile(str)
	if issues != nil && issues.Err() != nil {
		return nil, issues.Err()
	}
	pro, err := defaultEnv.Program(ast)
	if err != nil {
		return nil, err
	}

	out, _, err := pro.Eval(varsMap)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RegisterFunc 注册方法
func RegisterFunc(opts ...cel.EnvOption) {
	celConf.Lock()
	defer celConf.Unlock()
	celConf.celFunc = append(celConf.celFunc, opts...)
}

type Student struct {
	Age int32
	AgeRes string
}