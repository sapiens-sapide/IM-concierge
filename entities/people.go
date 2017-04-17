package entities

type Identity struct {
	Recipient
	UserId []byte `json:"user_id"`
}

type Recipient struct {
	DisplayName string `json:"display_name"`
	Identifier  string `json:"identifier"`
}
