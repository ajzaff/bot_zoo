package zoo

import (
	"regexp"
)

type Options struct {
	Verbose    bool
	Extensions bool
}

func (o *Options) SetVerbose(verbose bool) {
	o.Verbose = verbose
}

func (o *Options) EnableExtensions(extensions bool) {
	o.Extensions = extensions
}

func (o *Options) SetOption(name, value interface{}) error {
	return nil
}

var optionPattern = regexp.MustCompile(`^setoption name (\S+) value (\S+)$`)

func (e *Engine) ParseSetOption(text string) (name string, value interface{}, err error) {
	/*matches := optionPattern.FindStringSubmatch(strings.TrimSpace(text))
	if len(matches) != 3 {
		return "", "", fmt.Errorf("setoption does not match /%s/", optionPattern.String())
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
	case "hash":
		v, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		a.Logf("setting hash table size to %d MB", v)
		a.engine.table.Resize(v)

	// Unsupported for now:
	// 	"gused":
	// 	"sused":
	// 	"moveused":
	// 	"lastmoveused":
	// 	"depth":

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

	default:
		return fmt.Errorf("unsupported option: %q", option)
	}*/
	return "", nil, nil
}
