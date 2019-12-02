package template

import (
	"fmt"
	"go.starlark.net/resolve"
	"go.starlark.net/starlark"
	"log"
)

func init() {
	resolve.AllowFloat = true
	resolve.AllowLambda = true
	resolve.AllowNestedDef = true
	resolve.AllowRecursion = true
	resolve.AllowBitwise = true
}

type context map[string]starlark.Value

func newContext(data map[string]interface{}) (context, error) {
	ctx := make(context)
	for k, v := range data {
		val, err := newValue(v)
		if err != nil {
			return nil, err
		}
		ctx[k] = val
	}
	return ctx, nil
}

func (ctx context) toDict() (*starlark.Dict, error) {
	dict := starlark.NewDict(len(ctx))
	for k, v := range ctx {
		err := dict.SetKey(starlark.String(k), v)
		if err != nil {
			return nil, err
		}
	}
	return dict, nil
}

func (ctx context) toGoMap() (map[string]interface{}, error) {
	m := make(map[string]interface{})
	for k, v := range ctx {
		goval, err := parseValue(v)
		if err != nil {
			return nil, err
		}
		m[k] = goval
	}
	return m, nil
}

func execute(content string, ctx context) (starlark.Value, error) {
	thread := &starlark.Thread{
		Print: func(_ *starlark.Thread, msg string) { log.Println(msg) },
	}
	dict, err := ctx.toDict()
	if err != nil {
		return nil, err
	}
	predeclared := starlark.StringDict{"vars": dict}
	globals, err := starlark.ExecFile(thread, "", wrapScript(content), predeclared)
	if err != nil {
		//if evalErr, ok := err.(*starlark.EvalError); ok {
		//	log.(evalErr.Backtrace())
		//}
		return nil, err
	}
	val, err := starlark.Call(thread, globals["impl"], nil, nil)
	if err != nil {
		return nil, err
	}
	return val, nil
}

func wrapScript(c string) string {
	return fmt.Sprintf("def impl():\n%s\n", c)
}

func wrapInlineScript(c string) string {
	return fmt.Sprintf("\treturn %s", c)
}

func newValue(goval interface{}) (v starlark.Value, _ error) {
	switch eval := goval.(type) {
	case int:
		v = starlark.MakeInt(eval)
	case string:
		v = starlark.String(eval)
	case bool:
		v = starlark.Bool(eval)
	case []string:
		list := make([]starlark.Value, len(eval))
		for i, s := range eval {
			list[i] = starlark.String(s)
		}
		v = starlark.NewList(list)
	default:
		return nil, fmt.Errorf("go type %T currently is not supported", eval)
	}
	return v, nil
}

func parseValue(v starlark.Value) (interface{}, error) {
	switch val := v.(type) {
	case starlark.NoneType:
		return "", nil
	case starlark.Int:
		i, err := starlark.AsInt32(val)
		if err != nil {
			return nil, err
		}
		return i, nil
	case starlark.String:
		return string(val), nil
	case starlark.Bool:
		return bool(val), nil
	case starlark.Float:
		return float64(val), nil
	case starlark.Tuple, *starlark.List:
		list := make([]interface{}, val.(starlark.Indexable).Len())
		var err error
		for i := range list {
			list[i], err = parseValue(val.(starlark.Indexable).Index(i))
			if err != nil {
				return nil, err
			}
		}
		return list, nil
	case *starlark.Dict:
		dict := make(map[string]interface{}, val.Len())
		for _, k := range val.Keys() {
			v, _, err := val.Get(k)
			if err != nil {
				return nil, err
			}
			key, err := parseDictKey(k)
			if err != nil {
				return nil, err
			}
			dict[key], err = parseValue(v)
			if err != nil {
				return nil, err
			}
		}
		return dict, nil
	}
	return nil, fmt.Errorf("return type %q currently is not supported", v.Type())
}

func parseString(v starlark.Value) (string, error) {
	switch val := v.(type) {
	case starlark.NoneType:
		return "", nil
	case starlark.String:
		return string(val), nil
	}
	return "", fmt.Errorf("expected a string, got %s", v.Type())
}

func parseBool(v starlark.Value) bool {
	return bool(v.Truth())
}

func parseDictKey(v starlark.Value) (string, error) {
	switch val := v.(type) {
	case starlark.String:
		return string(val), nil
	case starlark.Int, starlark.Float, starlark.Bool, starlark.NoneType:
		return val.String(), nil
	case starlark.Tuple:
		var key string
		for i := 0; i < val.Len(); i++ {
			s, err := parseDictKey(val.Index(i))
			if err != nil {
				return "", err
			}
			if i != 0 {
				key += " "
			}
			key += s
		}
		return key, nil
	}
	return "", fmt.Errorf("data type %q is not supported as dictionary key", v.Type())
}
