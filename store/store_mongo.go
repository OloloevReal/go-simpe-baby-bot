package store

import "context"

type Store interface {
	Put(context.Context) error
}
