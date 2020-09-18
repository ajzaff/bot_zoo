package zoo

import "flag"

type AEISettings struct {
	BotName            string
	BotVersion         string
	BotAuthor          string
	LogProtocolTraffic bool
	Extensions         bool
	ProtoVersion       string
}

func RegisterAEIFlags(flagset *flag.FlagSet) *AEISettings {
	s := new(AEISettings)
	flagset.StringVar(&s.BotName, "bot_name", "bot_alpha_zoo", "Bot name reported by AEI `id'")
	flagset.StringVar(&s.BotVersion, "bot_version", "0", "Bot verion reported by AEI `id'")
	flagset.StringVar(&s.BotAuthor, "bot_author", "Alan Zaffetti", "Bot author reported by AEI `id'")
	flagset.BoolVar(&s.LogProtocolTraffic, "log_aei_protocol_traffic", false, "Log all AEI protocol messages sent and received to stderr")
	return s
}
