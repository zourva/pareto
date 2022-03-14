package box

import "github.com/pborman/uuid"


func UUID() string {
	return uuid.New()
}
