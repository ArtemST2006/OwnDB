package schema

import "time"

type Create struct {
	Name      string    `json: "name" binding: "required"`
	Timestamp time.Time `json: "time"`
}
