package db

import (
	"context"

	"github.com/VatsalP117/iris/pkg/core"
)

func (r *SqliteRepository) GetSystemStatus(ctx context.Context) (*core.SystemStatus, error) {
	if err := r.db.PingContext(ctx); err != nil {
		return nil, err
	}
	status := core.SystemStatus{Database: "ok"}
	if err := r.db.QueryRowContext(ctx, `
		SELECT
			COALESCE((SELECT last_seq FROM projection_checkpoints WHERE name = ?), 0),
			COALESCE((SELECT version FROM projection_checkpoints WHERE name = ?), 0),
			COALESCE((SELECT MAX(seq) FROM events), 0)
	`, analyticsProjectionName, analyticsProjectionName).Scan(
		&status.ProjectionLastSeq, &status.ProjectionVersion, &status.EventLastSeq,
	); err != nil {
		return nil, err
	}
	status.ProjectionLag = status.EventLastSeq - status.ProjectionLastSeq
	if status.ProjectionLag < 0 {
		status.ProjectionLag = 0
	}
	return &status, nil
}
