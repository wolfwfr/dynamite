package dynamodb

import (
	apitypes "github.com/wolfwfr/dynamite/pkg/adapters/dynamodb/types"
)

type options struct {
	objectStyling apitypes.ObjectStyling
}

type Option func(*options)

func WithObjectStyling(s apitypes.ObjectStyling) Option {
	return func(opts *options) {
		opts.objectStyling = s
	}
}
