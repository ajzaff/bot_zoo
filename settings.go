package zoo

import (
	"flag"
)

type AEISettings struct {
	BotName            string
	BotVersion         string
	BotAuthor          string
	LogProtocolTraffic bool
	LogVerbosePosition bool
	Extensions         bool
	ProtoVersion       string
}

func RegisterAEIFlags(flagset *flag.FlagSet) *AEISettings {
	s := new(AEISettings)
	flagset.StringVar(&s.ProtoVersion, "protocol_version", "1", "AEI protocol-version.")
	flagset.StringVar(&s.BotName, "bot_name", "bot_alpha_zoo", "Bot name reported by AEI `id`.")
	flagset.StringVar(&s.BotVersion, "bot_version", "0", "Bot verion reported by AEI `id`.")
	flagset.StringVar(&s.BotAuthor, "bot_author", "Alan Zaffetti", "Bot author reported by AEI `id`.")
	flagset.BoolVar(&s.LogProtocolTraffic, "log_protocol_traffic", false, "Log all AEI protocol messages sent and received to stderr.")
	flagset.BoolVar(&s.LogVerbosePosition, "log_verbose_position", false, "Log verbose position before all sent messages to stderr.")
	return s
}

type EngineSettings struct {
	UseTranspositionTable bool
	UsePonder             bool
	Seed                  int64
	MoveList              string
	Concurrency           uint
}

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
	return s
}
