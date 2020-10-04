package zoo

import (
	"flag"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// AEISettings contains AEI settings to configure the game engine.
// Usually bound to the current commandline using RegisterAEIFlags.
type AEISettings struct {
	BotName            string
	BotVersion         string
	BotAuthor          string
	LogProtocolTraffic bool
	Extensions         bool
	ProtoVersion       string
}

// RegisterAEIFlags registers engine flags to flagset.
func RegisterAEIFlags(flagset *flag.FlagSet) *AEISettings {
	s := new(AEISettings)
	flagset.StringVar(&s.ProtoVersion, "protocol_version", "1", "AEI protocol-version.")
	flagset.StringVar(&s.BotName, "bot_name", "bot_alpha_zoo", "Bot name reported by AEI `id`.")
	flagset.StringVar(&s.BotVersion, "bot_version", "0", "Bot verion reported by AEI `id`.")
	flagset.StringVar(&s.BotAuthor, "bot_author", "Alan Zaffetti", "Bot author reported by AEI `id`.")
	flagset.BoolVar(&s.LogProtocolTraffic, "log_protocol_traffic", false, "Log all AEI protocol messages sent and received to stderr.")
	return s
}

// EngineSettings contains engine settings to configure the game engine.
// Usually bound to the current commandline using RegisterEngineFlags.
type EngineSettings struct {
	UseTranspositionTable bool
	UsePonder             bool
	Seed                  int64
	MoveList              string
	Concurrency           uint
	UseDatasetWriter      bool
	DatasetEpoch          int
	PlayBatchGames        int
	UseSampledMove        bool
	Options               SetoptionFlag
}

// SetoptionElem is an element of the SetoptionFlag containing a single option and value override.
type SetoptionElem struct {
	Name   string
	StrVal string
}

var setoptionElemPattern = regexp.MustCompile(`^([\w]+)=("(?:[^"\\]|\\\w)*"|[\w.]+)$`)

func makeSetoptionElem(s string) (e SetoptionElem, err error) {
	match := setoptionElemPattern.FindStringSubmatch(s)
	if len(match) < 3 {
		return SetoptionElem{}, fmt.Errorf("failed to parse setoption flag: %q", s)
	}
	e.Name = match[1]
	strVal := match[2]
	if strings.HasPrefix(strVal, `"`) {
		strVal, err = strconv.Unquote(strVal)
		if err != nil {
			return SetoptionElem{}, fmt.Errorf("failed to unquote setoption flag value: %q: %v", match[2], err)
		}
	}
	e.StrVal = strVal
	return e, nil
}

// Execute executes setoption on o.
func (e SetoptionElem) Execute(o *Options) error {
	return o.ExecuteSetOption(fmt.Sprintf("name %s value %s", e.Name, e.StrVal))
}

func (e SetoptionElem) String() string {
	return fmt.Sprintf("%s=%q", e.Name, e.StrVal)
}

// SetoptionFlag is a custom flag type providing a slice of option names to override.
type SetoptionFlag []SetoptionElem

// Execute applies the setoption element e to o.
func (o *SetoptionFlag) Execute(opts *Options) error {
	for _, e := range *o {
		if err := e.Execute(opts); err != nil {
			return err
		}
	}
	return nil
}

// Set parses the setoption element appends it to the flag value or returns an error.
func (o *SetoptionFlag) Set(value string) error {
	e, err := makeSetoptionElem(value)
	if err != nil {
		return err
	}

	*o = append(*o, e)
	return nil
}

func (o *SetoptionFlag) String() string {
	var sb strings.Builder
	for _, e := range *o {
		sb.WriteString(e.String())
	}
	return sb.String()
}

// RegisterEngineFlags registers engine flags to flagset.
func RegisterEngineFlags(flagset *flag.FlagSet) *EngineSettings {
	s := new(EngineSettings)
	flagset.BoolVar(&s.UseTranspositionTable, "use_transposition_table", true, "Use the transposition table during search.")
	flagset.BoolVar(&s.UsePonder, "use_ponder", true, `Ponder on the opponent's turn.
Ponder implies we will search until we're asked explicitly to stop unless a terminal score is guaranteed.
We don't print the best move after a ponder. We don't clear the transposition table when we're done.`)
	flag.Int64Var(&s.Seed, "seed", 0, "Seed for random state in the engine. Defaults to a time-based seed.")
	flag.UintVar(&s.Concurrency, "concurrency", 1, "Use this number of goroutines for concurrent MCTS.")
	flag.StringVar(&s.MoveList, "movelist", "",
		`Execute "newgame" followed by the given movelist.
This should be the path to a move list file.  Each line should be prefixed by move number and color (e.g. 5s Rd4e).
Setup moves must be included. The last line may include a move number and color to indice the side to move.
The file may be newline terminated.`)
	flag.Var(&s.Options, "O", `Repeated flag used to set AEI options (e.g. -O foo=1 -O bar="xxx"`)
	flag.BoolVar(&s.UseDatasetWriter, "use_dataset_writer", false, "Enables the Dataset writer for outputting training data")
	flag.IntVar(&s.DatasetEpoch, "dataset_epoch", 0, "Epoch number to use when writing Dataset files")
	flag.IntVar(&s.PlayBatchGames, "playbatch_games", 5000, "Number of games to play for `playbatch'")
	flag.BoolVar(&s.UseSampledMove, "use_suboptimal_move", false, "Sample to best move instead of selecting the best")
	return s
}
