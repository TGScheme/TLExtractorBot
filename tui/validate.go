package tui

import (
	"TLExtractor/tui/types"
	"fmt"
	"net/url"
	"strconv"
)

func Validate(fieldName string, validate types.ValidateType) func(value string) error {
	return func(value string) error {
		if instance.isBack {
			return nil
		}
		if len(value) == 0 {
			return fmt.Errorf("%s is required", fieldName)
		}
		isValid := true
		defaultMessage := "%s is invalid"
		switch validate {
		case types.IsInt:
			if _, err := strconv.Atoi(value); err != nil {
				isValid = false
			}
		case types.IsFloat:
			if _, err := strconv.ParseFloat(value, 64); err != nil {
				isValid = false
			}
		case types.IsBool:
			if _, err := strconv.ParseBool(value); err != nil {
				isValid = false
			}
		case types.IsURL:
			defaultMessage = "%s is not a valid URL"
			u, err := url.Parse(value)
			isValid = err == nil && u.Scheme != "" && u.Host != ""
		case types.NoCheck:
		default:
			return nil
		}
		if !isValid {
			return fmt.Errorf(defaultMessage, fieldName)
		}
		return nil
	}
}
