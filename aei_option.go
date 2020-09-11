package zoo

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
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
		a.engine.timeControl.Move = time.Duration(v) * time.Second
	case "tcreserve":
		v, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		a.engine.timeControl.Reserve = time.Duration(v) * time.Second
	case "tcpercent":
		v, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		if v < 0 || v > 100 {
			return fmt.Errorf("percentage out of range [0-100%%]: %d", v)
		}
		a.engine.timeControl.MoveReservePercent = v
	case "tcmax":
		v, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		a.engine.timeControl.MaxReserve = time.Duration(v) * time.Second
	case "tctotal":
		v, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		a.engine.timeControl.GameTotal = time.Duration(v) * time.Second
	case "tcturns":
		v, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		a.engine.timeControl.Turns = v
	case "tcturntime":
		v, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		a.engine.timeControl.MaxTurn = time.Duration(v) * time.Second
	case "greserve":
		v, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		a.engine.timeInfo.Reserve[Gold] = time.Duration(v) * time.Second
	case "sreserve":
		v, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		a.engine.timeInfo.Reserve[Silver] = time.Duration(v) * time.Second
	// case "gused":
	// 	v, err := strconv.Atoi(value)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	a.engine.timeInfo. = time.Duration(v) * time.Second
	// case "sused":
	// case "moveused":
	// case "lastmoveused":
	case "hash":
		v, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		n := 1e6 * v / EntrySize
		a.Logf("setting hash table size to %d entries (%d MiB)", n, v)
		a.engine.table.SetCap(n)
	case "depth":
		v, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		if v < 0 {
			return fmt.Errorf("depth < 0")
		}
		a.engine.fixedDepth = v

		// Custom Zoo engine options:
	case "goroutines":
		v, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		if v < 0 {
			return fmt.Errorf("goroutines <= 0")
		}
		a.engine.concurrency = v
	case "mindepth":
		v, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		if v < 0 {
			return fmt.Errorf("depth < 0")
		}
		a.engine.minDepth = v

	default:
		return fmt.Errorf("unsupported option: %q", option)
	}
	return nil
}
