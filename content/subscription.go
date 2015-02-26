package content

import (
	"fmt"

	"github.com/urandom/readeef/content/info"
)

type Subscription interface {
	Error
	RepoRelated

	fmt.Stringer

	Info(in ...info.Subscription) info.Subscription

	Validate() error

	Update()
	Delete()
}