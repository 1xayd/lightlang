package builtins

import (
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type BuiltinFunc func(args []interface{}) (interface{}, error)

func toFloat64(val interface{}) float64 {
	if v, ok := val.(float64); ok {
		return v
	}
	if v, ok := val.(int); ok {
		return float64(v)
	}
	if v, ok := val.(int64); ok {
		return float64(v)
	}
	if v, ok := val.(int32); ok {
		return float64(v)
	}
	return 0.0
}

var Builtins = map[string]BuiltinFunc{
	"print": func(args []interface{}) (interface{}, error) {
		fmt.Print(fmt.Sprintln(args...))
		return nil, nil
	},

	"range": func(args []interface{}) (interface{}, error) {
		switch len(args) {
		case 1:
			end := int(toFloat64(args[0]))
			result := make([]interface{}, end)
			for i := 0; i < end; i++ {
				result[i] = float64(i)
			}
			return result, nil
		case 2:
			start := int(toFloat64(args[0]))
			end := int(toFloat64(args[1]))
			result := make([]interface{}, end-start)
			for i := start; i < end; i++ {
				result[i-start] = float64(i)
			}
			return result, nil
		case 3:
			start := int(toFloat64(args[0]))
			end := int(toFloat64(args[1]))
			step := int(toFloat64(args[2]))
			if step == 0 {
				return nil, fmt.Errorf("range step cannot be 0")
			}

			size := 0
			if step > 0 {
				size = (end - start + step - 1) / step
			} else {
				size = (start - end - step - 1) / -step
			}

			if size < 0 {
				size = 0
			}

			result := make([]interface{}, size)
			for i := 0; i < size; i++ {
				result[i] = float64(start + i*step)
			}
			return result, nil
		default:
			return nil, fmt.Errorf("range expects 1-3 arguments")
		}
	},

	"pairs": func(args []interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("pairs expects 1 argument")
		}

		switch v := args[0].(type) {
		case map[string]interface{}:
			pairs := make([]interface{}, 0, len(v))
			for key, val := range v {
				pair := []interface{}{key, val}
				pairs = append(pairs, pair)
			}
			return pairs, nil
		case []interface{}:
			pairs := make([]interface{}, len(v))
			for i, val := range v {
				pairs[i] = []interface{}{float64(i), val}
			}
			return pairs, nil
		default:
			return nil, fmt.Errorf("pairs requires table or array")
		}
	},

	"ipairs": func(args []interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("ipairs expects 1 argument")
		}

		switch v := args[0].(type) {
		case []interface{}:
			pairs := make([]interface{}, 0, len(v))
			for i, val := range v {
				pairs = append(pairs, []interface{}{float64(i + 1), val})
			}
			return pairs, nil
		default:
			return nil, fmt.Errorf("ipairs requires array")
		}
	},

	"len": func(args []interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("len expects 1 argument")
		}
		switch v := args[0].(type) {
		case []interface{}:
			return float64(len(v)), nil
		case map[string]interface{}:
			return float64(len(v)), nil
		case string:
			return float64(len(v)), nil
		default:
			return nil, fmt.Errorf("len invalid type")
		}
	},

	"type": func(args []interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("type expects 1 argument")
		}
		return fmt.Sprintf("%T", args[0]), nil
	},

	"push": func(args []interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("push expects 2 arguments (table, value)")
		}
		switch tbl := args[0].(type) {
		case []interface{}:
			return append(tbl, args[1]), nil
		default:
			return nil, fmt.Errorf("push requires array")
		}
	},

	"sqrt": func(args []interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("sqrt expects 1 argument")
		}
		if f, ok := args[0].(float64); ok {
			return math.Sqrt(f), nil
		}
		return nil, fmt.Errorf("sqrt requires number")
	},

	"abs": func(args []interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("abs expects 1 argument")
		}
		if f, ok := args[0].(float64); ok {
			return math.Abs(f), nil
		}
		return nil, fmt.Errorf("abs requires number")
	},

	"pow": func(args []interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("pow expects 2 arguments (base, exponent)")
		}
		base, ok1 := args[0].(float64)
		exp, ok2 := args[1].(float64)
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("pow requires numbers")
		}
		return math.Pow(base, exp), nil
	},

	"sin": func(args []interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("sin expects 1 argument")
		}
		if f, ok := args[0].(float64); ok {
			return math.Sin(f), nil
		}
		return nil, fmt.Errorf("sin requires number")
	},

	"cos": func(args []interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("cos expects 1 argument")
		}
		if f, ok := args[0].(float64); ok {
			return math.Cos(f), nil
		}
		return nil, fmt.Errorf("cos requires number")
	},

	"tan": func(args []interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("tan expects 1 argument")
		}
		if f, ok := args[0].(float64); ok {
			return math.Tan(f), nil
		}
		return nil, fmt.Errorf("tan requires number")
	},

	"log": func(args []interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("log expects 1 argument")
		}
		if f, ok := args[0].(float64); ok {
			return math.Log(f), nil
		}
		return nil, fmt.Errorf("log requires number")
	},

	"exp": func(args []interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("exp expects 1 argument")
		}
		if f, ok := args[0].(float64); ok {
			return math.Exp(f), nil
		}
		return nil, fmt.Errorf("exp requires number")
	},

	"floor": func(args []interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("floor expects 1 argument")
		}
		if f, ok := args[0].(float64); ok {
			return math.Floor(f), nil
		}
		return nil, fmt.Errorf("floor requires number")
	},

	"clamp": func(args []interface{}) (interface{}, error) {
		if len(args) != 3 {
			return nil, fmt.Errorf("clamp expects 3 arguments (value, min, max)")
		}

		val, ok1 := args[0].(float64)
		min, ok2 := args[1].(float64)
		max, ok3 := args[2].(float64)
		if !ok1 || !ok2 || !ok3 {
			return nil, fmt.Errorf("clamp requires numbers")
		}

		if val < min {
			return min, nil
		}
		if val > max {
			return max, nil
		}
		return val, nil
	},

	"lerp": func(args []interface{}) (interface{}, error) {
		if len(args) != 3 {
			return nil, fmt.Errorf("lerp expects 3 arguments (a, b, t)")
		}

		a, ok1 := args[0].(float64)
		b, ok2 := args[1].(float64)
		t, ok3 := args[2].(float64)
		if !ok1 || !ok2 || !ok3 {
			return nil, fmt.Errorf("lerp requires numbers")
		}

		return a + t*(b-a), nil
	},

	"ceil": func(args []interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("ceil expects 1 argument")
		}
		if f, ok := args[0].(float64); ok {
			return math.Ceil(f), nil
		}
		return nil, fmt.Errorf("ceil requires number")
	},

	"round": func(args []interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("round expects 1 argument")
		}
		if f, ok := args[0].(float64); ok {
			return math.Round(f), nil
		}
		return nil, fmt.Errorf("round requires number")
	},

	"max": func(args []interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("max expects at least 1 argument")
		}
		maxVal := math.Inf(-1)
		for _, arg := range args {
			if f, ok := arg.(float64); ok {
				if f > maxVal {
					maxVal = f
				}
			} else {
				return nil, fmt.Errorf("max requires numbers")
			}
		}
		return maxVal, nil
	},

	"min": func(args []interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("min expects at least 1 argument")
		}
		minVal := math.Inf(1)
		for _, arg := range args {
			if f, ok := arg.(float64); ok {
				if f < minVal {
					minVal = f
				}
			} else {
				return nil, fmt.Errorf("min requires numbers")
			}
		}
		return minVal, nil
	},

	"substr": func(args []interface{}) (interface{}, error) {
		if len(args) != 3 {
			return nil, fmt.Errorf("substr expects 3 arguments (string, start, length)")
		}
		str, ok1 := args[0].(string)
		start, ok2 := args[1].(float64)
		length, ok3 := args[2].(float64)
		if !ok1 || !ok2 || !ok3 {
			return nil, fmt.Errorf("substr requires (string, number, number)")
		}

		s := int(start)
		l := int(length)
		if s < 0 || s >= len(str) || l < 0 {
			return "", nil
		}
		if s+l > len(str) {
			l = len(str) - s
		}
		return str[s : s+l], nil
	},

	"concat": func(args []interface{}) (interface{}, error) {
		result := ""
		for _, arg := range args {
			result += fmt.Sprintf("%v", arg)
		}
		return result, nil
	},

	"upper": func(args []interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("upper expects 1 argument")
		}
		if s, ok := args[0].(string); ok {
			return strings.ToUpper(s), nil
		}
		return nil, fmt.Errorf("upper requires string")
	},

	"lower": func(args []interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("lower expects 1 argument")
		}
		if s, ok := args[0].(string); ok {
			return strings.ToLower(s), nil
		}
		return nil, fmt.Errorf("lower requires string")
	},

	"split": func(args []interface{}) (interface{}, error) {
		if len(args) != 1 && len(args) != 2 {
			return nil, fmt.Errorf("split expects 1 or 2 arguments")
		}
		if s, ok := args[0].(string); ok {
			sep := " "
			if len(args) == 2 {
				if sepStr, ok := args[1].(string); ok {
					sep = sepStr
				} else {
					return nil, fmt.Errorf("split separator must be string")
				}
			}
			parts := strings.Split(s, sep)
			result := make([]interface{}, len(parts))
			for i, p := range parts {
				result[i] = p
			}
			return result, nil
		}
		return nil, fmt.Errorf("split requires string")
	},

	"find": func(args []interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("find expects 2 arguments (string, substring)")
		}
		if s, ok1 := args[0].(string); ok1 {
			if sub, ok2 := args[1].(string); ok2 {
				index := strings.Index(s, sub)
				return float64(index), nil
			}
		}
		return nil, fmt.Errorf("find requires strings")
	},

	"replace": func(args []interface{}) (interface{}, error) {
		if len(args) != 3 {
			return nil, fmt.Errorf("replace expects 3 arguments (string, old, new)")
		}
		if s, ok1 := args[0].(string); ok1 {
			if old, ok2 := args[1].(string); ok2 {
				if new, ok3 := args[2].(string); ok3 {
					return strings.ReplaceAll(s, old, new), nil
				}
			}
		}
		return nil, fmt.Errorf("replace requires strings")
	},

	"pop": func(args []interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("pop expects 1 argument (array)")
		}
		switch arr := args[0].(type) {
		case []interface{}:
			if len(arr) == 0 {
				return arr, nil
			}
			return arr[:len(arr)-1], nil
		default:
			return nil, fmt.Errorf("pop requires array")
		}
	},

	"keys": func(args []interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("keys expects 1 argument")
		}
		switch m := args[0].(type) {
		case map[string]interface{}:
			keys := make([]interface{}, 0, len(m))
			for k := range m {
				keys = append(keys, k)
			}
			return keys, nil
		default:
			return nil, fmt.Errorf("keys requires map")
		}
	},

	"tick": func(args []interface{}) (interface{}, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("tick expects 0 arguments")
		}
		now := time.Now()
		return float64(now.Unix()) + float64(now.Nanosecond())/1e9, nil
	},

	"time": func(args []interface{}) (interface{}, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("time expects 0 arguments")
		}
		return float64(time.Now().Unix()), nil
	},

	"date": func(args []interface{}) (interface{}, error) {
		now := time.Now()
		if len(args) == 0 {
			return map[string]interface{}{
				"year":  float64(now.Year()),
				"month": float64(now.Month()),
				"day":   float64(now.Day()),
				"hour":  float64(now.Hour()),
				"min":   float64(now.Minute()),
				"sec":   float64(now.Second()),
				"wday":  float64(now.Weekday()),
				"yday":  float64(now.YearDay()),
				"isdst": now.IsDST(),
				"epoch": float64(now.Unix()),
				"tick":  float64(now.Unix()) + float64(now.Nanosecond())/1e9,
			}, nil
		}

		if len(args) == 1 {
			if ts, ok := args[0].(float64); ok {
				seconds := int64(ts)
				nanoseconds := int64((ts - float64(seconds)) * 1e9)
				t := time.Unix(seconds, nanoseconds)
				return map[string]interface{}{
					"year":  float64(t.Year()),
					"month": float64(t.Month()),
					"day":   float64(t.Day()),
					"hour":  float64(t.Hour()),
					"min":   float64(t.Minute()),
					"sec":   float64(t.Second()),
					"wday":  float64(t.Weekday()),
					"yday":  float64(t.YearDay()),
					"isdst": t.IsDST(),
					"epoch": float64(t.Unix()),
					"tick":  ts,
				}, nil
			}
			return nil, fmt.Errorf("date requires number or no arguments")
		}

		return nil, fmt.Errorf("date expects 0 or 1 argument")
	},

	"wait": func(args []interface{}) (interface{}, error) {
		var seconds float64 = 0
		if len(args) == 1 {
			if s, ok := args[0].(float64); ok {
				seconds = s
			} else {
				return nil, fmt.Errorf("wait requires number")
			}
		} else if len(args) > 1 {
			return nil, fmt.Errorf("wait expects 0 or 1 argument")
		}

		duration := time.Duration(seconds * float64(time.Second))
		time.Sleep(duration)
		return seconds, nil
	},

	"random": func(args []interface{}) (interface{}, error) {
		if len(args) > 2 {
			return nil, fmt.Errorf("random expects 0, 1, or 2 arguments")
		}

		if len(args) == 0 {
			return float64(rand.Intn(1 - 0)), nil
		}

		if len(args) == 1 {
			if max, ok := args[0].(float64); ok {
				if max <= 0 {
					return nil, fmt.Errorf("random max must be positive")
				}
				return float64(rand.Intn(int(max))), nil
			}
			return nil, fmt.Errorf("random requires number")
		}

		min, ok1 := args[0].(float64)
		max, ok2 := args[1].(float64)
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("random requires numbers")
		}
		if max <= min {
			return nil, fmt.Errorf("random max must be greater than min")
		}
		return min + float64(rand.Intn(int(max-min))), nil
	},

	"tostring": func(args []interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("tostring() expects 1 argument")
		}
		return fmt.Sprintf("%v", args[0]), nil
	},

	"tonumber": func(args []interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("tonumber() expects 1 argument")
		}
		switch v := args[0].(type) {
		case float64:
			return v, nil
		case string:
			var f float64
			_, err := fmt.Sscanf(v, "%f", &f)
			if err != nil {
				return nil, fmt.Errorf("cannot convert string to number")
			}
			return f, nil
		default:
			return nil, fmt.Errorf("cannot convert to number")
		}
	},
	"writefile": func(args []interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("writefile expects 2 arguments (filename, content)")
		}

		filename, ok1 := args[0].(string)
		if !ok1 {
			return nil, fmt.Errorf("writefile filename must be string")
		}

		content := fmt.Sprintf("%v", args[1])

		dir := filepath.Dir(filename)
		if dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create directory: %v", err)
			}
		}

		err := ioutil.WriteFile(filename, []byte(content), 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to write file: %v", err)
		}

		return nil, nil
	},

	"readfile": func(args []interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("readfile expects 1 argument (filename)")
		}

		filename, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("readfile filename must be string")
		}

		data, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %v", err)
		}

		return string(data), nil
	},

	"makedir": func(args []interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("makedir expects 1 argument (dirname)")
		}

		dirname, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("makedir dirname must be string")
		}

		err := os.MkdirAll(dirname, 0755)
		if err != nil {
			return nil, fmt.Errorf("failed to create directory: %v", err)
		}

		return nil, nil
	},

	"gotodir": func(args []interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("gotodir expects 1 argument (dirname)")
		}

		dirname, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("gotodir dirname must be string")
		}

		info, err := os.Stat(dirname)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("directory does not exist: %s", dirname)
			}
			return nil, fmt.Errorf("failed to access directory: %v", err)
		}

		if !info.IsDir() {
			return nil, fmt.Errorf("not a directory: %s", dirname)
		}

		err = os.Chdir(dirname)
		if err != nil {
			return nil, fmt.Errorf("failed to change directory: %v", err)
		}

		return nil, nil
	},
}
