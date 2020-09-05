package zoo

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
	"unsafe"
)

var optionPattern = regexp.MustCompile(`^setoption name (\S+) value (\S+)$`)

func (a *AEI) handleOption(text string) error {
	matches := optionPattern.FindStringSubmatch(text)
	if len(matches) != 3 {
		return fmt.Errorf("setoption does not match /%s/", optionPattern.String())
	}
	switch option, value := matches[1], matches[2]; option {
	case "tcmove":
		v, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		a.engine.TimeLimits.Move = time.Duration(v) * time.Second
	case "tcreserve":
		v, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		a.engine.TimeLimits.Reserve = time.Duration(v) * time.Second
	case "tcpercent":
		v, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		if v < 0 || v > 100 {
			return fmt.Errorf("percentage out of range [0-100%]: %d", v)
		}
		a.engine.TimeLimits.Percent = v
	case "tcmax":
		v, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		a.engine.TimeLimits.Max = time.Duration(v) * time.Second
	case "tctotal":
		v, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		a.engine.TimeLimits.Total = time.Duration(v) * time.Second
	case "tcturns":
		v, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		a.engine.TimeLimits.Turns = v
	case "tcturntime":
		v, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		a.engine.TimeLimits.Turn = time.Duration(v) * time.Second
	case "greserve":
		v, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		a.engine.TimeInfo.Reserve[Gold] = time.Duration(v) * time.Second
	case "sreserve":
		v, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		a.engine.TimeInfo.Reserve[Silver] = time.Duration(v) * time.Second
	case "gused":
		v, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		a.engine.TimeInfo.Used[Gold] = time.Duration(v) * time.Second
	case "sused":
		v, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		a.engine.TimeInfo.Used[Silver] = time.Duration(v) * time.Second
	case "lastmoveused":
		v, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		a.engine.TimeInfo.LastMoveUsed = time.Duration(v) * time.Second
	case "moveused":
		v, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		a.engine.TimeInfo.MoveUsed = time.Duration(v) * time.Second
	case "hash":
		v, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		size := unsafe.Sizeof(TableEntry{})
		n := 1e6 * v / int(size)
		a.Logf("setting hash table size to %d entries (%d MiB)", n, v)
		a.engine.table = NewTable(n)
	case "depth":
		v, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		if v < 0 {
			return fmt.Errorf("depth < 0")
		}
		a.engine.depth = v
	default:
		return fmt.Errorf("unsupported option: %q", option)
	}
	return nil
}
