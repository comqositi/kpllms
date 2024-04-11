package kpllms

import (
	"context"
	"github.com/comqositi/kpllms/schema"
)

type Model interface {
	Chat(ctx context.Context, messages []*schema.ChatMessage, options ...CallOption) (*schema.ContentResponse, error)
}
