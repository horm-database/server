package conf

import (
	"reflect"
	"strings"
	"time"

	"github.com/horm-database/common/errs"
	"github.com/horm-database/common/types"
)

type FilterConfig map[string]interface{}

// GetBool 获取 bool 类型配置
func (f FilterConfig) GetBool(key string) (ret bool, exist bool) {
	return types.GetBool(f, key)
}

// GetString 获取 string 类型配置
func (f FilterConfig) GetString(key string) (ret string, exist bool) {
	return types.GetString(f, key)
}

// GetInt 获取 int 类型配置
func (f FilterConfig) GetInt(key string) (ret int64, exist bool, err error) {
	ret, exist, err = types.GetInt64(f, key)
	if err != nil {
		err = errs.Newf(errs.RetFilterConfigInvalid, "get filter int config error:", err)
	}
	return
}

// GetUint 获取 uint 类型配置
func (f FilterConfig) GetUint(key string) (ret uint64, exist bool, err error) {
	ret, exist, err = types.GetUint64(f, key)
	if err != nil {
		err = errs.Newf(errs.RetFilterConfigInvalid, "get filter uint config error:", err)
	}
	return
}

// GetFloat 获取 float 类型配置
func (f FilterConfig) GetFloat(key string) (ret float64, exist bool, err error) {
	ret, exist, err = types.GetFloat64(f, key)
	if err != nil {
		err = errs.Newf(errs.RetFilterConfigInvalid, "get filter float config error:", err)
	}
	return
}

// GetBytes 获取 bytes 类型配置
func (f FilterConfig) GetBytes(key string) (ret []byte, exist bool) {
	return types.GetBytes(f, key)
}

// GetEnum 获取枚举（单选）类型配置
func (f FilterConfig) GetEnum(key string) (ret string, exist bool) {
	return types.GetString(f, key)
}

// GetTime 获取 date、time 类型配置
func (f FilterConfig) GetTime(key string, loc ...*time.Location) (ret time.Time, exist bool, err error) {
	if len(f) == 0 {
		return time.Time{}, false, nil
	}

	value, ok := f[key]
	if !ok {
		return time.Time{}, false, nil
	}

	if value == nil {
		return time.Time{}, false, nil
	}

	switch val := value.(type) {
	case time.Time:
		return val, true, nil
	case *time.Time:
		if val == nil {
			return time.Time{}, true, nil
		} else {
			return *val, true, nil
		}
	default:
		t := strings.TrimSpace(types.InterfaceToString(value))

		l := time.Local
		if len(loc) > 0 {
			l = loc[0]
		}

		layout := "2006-01-02 15:04:05"
		if len(t) == 10 {
			layout = "2006-01-02"
		}

		ret, err = time.ParseInLocation(layout, t, l)
		if err != nil {
			err = errs.Newf(errs.RetFilterConfigInvalid, "get filter time config error:", err)
		}

		return ret, true, err
	}
}

// GetTimeInterval 获取 date、time 时间区间
func (f FilterConfig) GetTimeInterval(key string, loc ...*time.Location) (start, end time.Time, exist bool, err error) {
	if len(f) == 0 {
		return time.Time{}, time.Time{}, false, nil
	}

	value, ok := f[key]
	if !ok {
		return time.Time{}, time.Time{}, false, nil
	}

	if value == nil {
		return time.Time{}, time.Time{}, false, nil
	}

	times, ok := value.([]time.Time)
	if ok && len(times) == 2 {
		return times[0], times[1], true, nil
	}

	str := types.InterfaceToString(value)

	t := strings.Split(str, "~")
	if len(t) != 2 {
		return time.Time{}, time.Time{}, true, errs.New(errs.RetFilterConfigInvalid,
			"get date interval config error: config value should have start time and end time")
	}

	l := time.Local
	if len(loc) > 0 {
		l = loc[0]
	}

	startStr := strings.TrimSpace(t[0])
	if len(startStr) == 10 {
		start, err = types.InterfaceToTime(startStr, "2006-01-02", l)
	} else {
		start, err = types.InterfaceToTime(startStr, "2006-01-02 15:04:05", l)
	}

	if err != nil {
		return time.Time{}, time.Time{}, true,
			errs.Newf(errs.RetFilterConfigInvalid, "get filter time interval config start time error:", err)
	}

	endStr := strings.TrimSpace(t[1])
	if len(endStr) == 10 {
		end, err = types.InterfaceToTime(endStr, "2006-01-02", l)
	} else {
		end, err = types.InterfaceToTime(endStr, "2006-01-02 15:04:05", l)
	}

	if err != nil {
		return time.Time{}, time.Time{}, true,
			errs.Newf(errs.RetFilterConfigInvalid, "get filter time interval config end time error:", err)
	}

	return start, end, true, nil
}

// GetStringArray 获取 string 数组配置
func (f FilterConfig) GetStringArray(key string) (ret []string, exist bool, err error) {
	ret, exist, err = types.GetStringArray(f, key)
	if err != nil {
		err = errs.Newf(errs.RetFilterConfigInvalid, "get filter string array config error:", err)
	}
	return
}

// GetIntArray 获取 int 数组配置
func (f FilterConfig) GetIntArray(key string) (ret []int64, exist bool, err error) {
	ret, exist, err = types.GetInt64Array(f, key)
	if err != nil {
		err = errs.Newf(errs.RetFilterConfigInvalid, "get filter int array config error:", err)
	}
	return
}

// GetUintArray 获取 uint 数组配置
func (f FilterConfig) GetUintArray(key string) (ret []uint64, exist bool, err error) {
	ret, exist, err = types.GetUint64Array(f, key)
	if err != nil {
		err = errs.Newf(errs.RetFilterConfigInvalid, "get filter uint array config error:", err)
	}
	return
}

// GetFloatArray 获取 float 数组配置
func (f FilterConfig) GetFloatArray(key string) (ret []float64, exist bool, err error) {
	ret, exist, err = types.GetFloat64Array(f, key)
	if err != nil {
		err = errs.Newf(errs.RetFilterConfigInvalid, "get filter float array config error:", err)
	}
	return
}

// GetMapConf 获取 map 类型配置
func (f FilterConfig) GetMapConf(key string) (FilterConfig, bool, error) {
	tmp, exist, err := types.GetMap(f, key)
	if err != nil {
		return nil, exist, errs.Newf(errs.RetFilterConfigInvalid, "get filter map config error:", err)
	}

	return FilterConfig(tmp), exist, nil
}

// GetMultiConf 获取配置数组
func (f FilterConfig) GetMultiConf(key string) (ret []FilterConfig, exist bool, err error) {
	value, ok := f[key]
	if !ok {
		return nil, false, nil
	}

	if value == nil {
		return nil, true, nil
	}

	switch arrVal := value.(type) {
	case []FilterConfig:
		return arrVal, true, nil
	case []interface{}:
		ret = make([]FilterConfig, len(arrVal))
		for k, arrItem := range arrVal {
			im, e := types.InterfaceToMap(arrItem)
			if e != nil {
				return nil, true, errs.Newf(errs.RetFilterConfigInvalid, "get filter multi-conf error:", e)
			}
			ret[k] = im
		}
	case []map[string]interface{}:
		ret = make([]FilterConfig, len(arrVal))
		for k, arrItem := range arrVal {
			ret[k] = arrItem
		}
	default:
		v := reflect.ValueOf(value)
		if types.IsNil(v) {
			return nil, true, nil
		}

		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		if !types.IsArray(v) {
			return nil, true, errs.Newf(errs.RetFilterConfigInvalid, "get filter multi-conf error: value is not array")
		}

		l := v.Len()
		ret = make([]FilterConfig, l)

		for i := 0; i < l; i++ {
			im, e := types.InterfaceToMap(types.Interface(v.Index(i)))
			if e != nil {
				return nil, true, errs.Newf(errs.RetFilterConfigInvalid, "get filter multi-conf error:", e)
			}
			ret[i] = im
		}
	}

	return ret, true, nil
}
