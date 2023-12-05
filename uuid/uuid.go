package uuid

import "github.com/pborman/uuid"

// UUID returns a random (version 4) UUID as a string.
func UUID() string {
	return uuid.New()
}
