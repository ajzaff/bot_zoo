package zoo

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// Options is a thread-safe map that stores the parsed results of `setoption` AEI commands.
type Options struct {
	data map[string]interface{}
	m    sync.RWMutex
}

func newOptions() *Options {
	o := &Options{data: make(map[string]interface{})}
	o.ExecuteSetOption("name playouts value 1")
	return o
}

// GetOption returns the value of the named option.
func (o *Options) GetOption(name string) interface{} {
	o.m.RLock()
	defer o.m.RUnlock()
	return o.data[name]
}

// LookupOption returns the value of the named option and a bool indicating whether it existed.
func (o *Options) LookupOption(name string) (value interface{}, ok bool) {
	o.m.RLock()
	defer o.m.RUnlock()
	v, ok := o.data[name]
	return v, ok
}

func (o *Options) Range(f func(key string, value interface{})) {
	o.m.Lock()
	defer o.m.Unlock()

	var names []string
	for name := range o.data {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		f(name, o.data[name])
	}
}

var setOptionPattern = regexp.MustCompile(`^name (\S+) value (\S+)$`)

// ExecuteSetOption parses the setoption arguments and attempts to call the setoption handler or returns an error.
// Example input:
//	"name foo_option value 10"
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
	o.m.Lock()
	defer o.m.Unlock()
	o.data[name] = value
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
