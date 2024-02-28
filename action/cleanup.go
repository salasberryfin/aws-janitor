package action

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
)

type CleanupScope struct {
	Session *session.Session
	TTL     time.Duration
	Commit  bool
	Pattern string
}

type CleanupFunc func(ctx context.Context, input *CleanupScope) error
