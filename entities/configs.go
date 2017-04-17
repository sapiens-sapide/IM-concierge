package entities

type (
	AttachmentType uint8
	MessageType    uint8
	ProtocolType   uint8
	RecipientType  uint8
)

const (
	SMTPmail MessageType = iota
	IRCmsg
)
const (
	Image AttachmentType = iota
	URL
)
const (
	From RecipientType = iota
	To
	Cc
	Bcc
)

const (
	Email ProtocolType = iota
	IRC
	XMPP
)

type ConciergeConfig struct {
	IRCserver   string      `mapstructure:"irc_server"`
	IRCRoom     string      `mapstructure:"irc_room"`
	IRCUser     string      `mapstructure:"irc_user"`
	IRCNickname string      `mapstructure:"irc_nickname"`
	Backend     EliasConfig `mapstructure:"EliasConfig"`
}

type EliasConfig struct {
}
