package zoo

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Options is a map that stores the parsed results of `setoption` AEI commands.
type Options map[string]interface{}

var setOptionPattern = regexp.MustCompile(`^name (\S+) value (\S+)$`)

// ExecuteSetOption parses the setoption arguments and attempts to call the setoption handler or returns an error.
func (o *Options) ExecuteSetOption(s string) error {
	matches := setOptionPattern.FindStringSubmatch(strings.TrimSpace(s))
	if len(matches) != 3 {
		return fmt.Errorf("setoption does not match /%s/", setOptionPattern.String())
	}
	name, strVal := matches[1], matches[2]
	handler := globalSetOptions[name]
	if handler == nil {
		return fmt.Errorf("unrecognized option: %s", name)
	}
	value, err := handler(strVal)
	if err != nil {
		return err
	}
	(*o)[name] = value
	return nil
}

var globalSetOptions = make(map[string]func(strVal string) (value interface{}, err error))

// RegisterSetOption registers the handler function to be called on setoption commands matching the given option name.
func RegisterSetOption(name string, handler func(s string) (value interface{}, err error)) {
	globalSetOptions[name] = handler
}

func setIntOptionFunc() func(s string) (value interface{}, err error) {
	return func(s string) (value interface{}, err error) {
		v, err := strconv.Atoi(s)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

func init() {
	RegisterSetOption("tcmove", setIntOptionFunc())
	RegisterSetOption("tcreserve", setIntOptionFunc())
	RegisterSetOption("tcpercent", setIntOptionFunc())
	RegisterSetOption("tcmax", setIntOptionFunc())
	RegisterSetOption("tctotal", setIntOptionFunc())
	RegisterSetOption("tcturns", setIntOptionFunc())
	RegisterSetOption("tcturntime", setIntOptionFunc())
	RegisterSetOption("greserve", setIntOptionFunc())
	RegisterSetOption("sreserve", setIntOptionFunc())
	RegisterSetOption("hash", setIntOptionFunc())
	RegisterSetOption("goroutines", setIntOptionFunc())
}
