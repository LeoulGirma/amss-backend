package streams

import (
	"context"

	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const importJobsStream = "import.jobs"

type ImportQueue struct {
	Client *redis.Client
}

var _ ports.ImportJobQueue = (*ImportQueue)(nil)

func (q *ImportQueue) Enqueue(ctx context.Context, importID uuid.UUID) error {
	if q == nil || q.Client == nil {
		return nil
	}
	_, err := q.Client.XAdd(ctx, &redis.XAddArgs{
		Stream: importJobsStream,
		Values: map[string]any{
			"import_id": importID.String(),
		},
	}).Result()
	return err
}
