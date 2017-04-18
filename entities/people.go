package entities

import "github.com/satori/go.uuid"

type Identity struct {
	DisplayName string    `json:"display_name"`
	Identifier  string    `json:"identifier"`
	UserId      uuid.UUID `json:"user_id"`
}

type Recipient struct {
	DisplayName string `json:"display_name"`
	Identifier  string `json:"identifier"`
}
