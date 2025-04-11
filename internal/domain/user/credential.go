package user

import "time"

type Credential struct {
	ID              string `json:"id" bson:"_id"`
	Name            string `json:"name" bson:"name"`
	Description     string `json:"description" bson:"description"`
	Username        string `json:"username" bson:"username"`
	EncryptPassword string `json:"encrypt_password" bson:"encrypt_password"`
	URL             string `json:"url" bson:"url"`

	VaultID string `json:"vault_id" bson:"vault_id"`

	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}
