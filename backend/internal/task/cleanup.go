package task

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"stdoutcms/internal/storage"
	"stdoutcms/internal/store"
)

// CleanupOrphanImages returns a scheduler-compatible task that scans MinIO
// for images under "images/" and deletes any that are not referenced by any
// post's content.
func CleanupOrphanImages(logger *slog.Logger, pg *store.Postgres, s3 *storage.S3) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		// 1. List all images in MinIO
		keys, err := s3.ListObjectKeys(ctx, "images/")
		if err != nil {
			return fmt.Errorf("list objects: %w", err)
		}
		if len(keys) == 0 {
			logger.Info("orphan-images: no images found, nothing to do")
			return nil
		}

		// 2. Get all post content
		contents, err := pg.GetAllContent(ctx)
		if err != nil {
			return fmt.Errorf("get all content: %w", err)
		}

		// 3. Find orphans — keys not referenced in any post content
		orphans := findOrphans(keys, contents)

		if len(orphans) == 0 {
			logger.Info("orphan-images: no orphans found", "total", len(keys))
			return nil
		}

		logger.Info("orphan-images: deleting orphans", "orphans", len(orphans), "total", len(keys))

		// 4. Delete orphans
		var deleted int
		var errs []string
		for _, key := range orphans {
			if err := s3.Delete(ctx, key); err != nil {
				errs = append(errs, fmt.Sprintf("%s: %v", key, err))
			} else {
				deleted++
			}
		}

		logger.Info("orphan-images: cleanup finished",
			"deleted", deleted,
			"failed", len(errs),
			"total_images", len(keys),
		)

		if len(errs) > 0 {
			return fmt.Errorf("partial cleanup: %d/%d deleted, errors: %s",
				deleted, len(orphans), strings.Join(errs, "; "))
		}

		return nil
	}
}

// findOrphans returns keys that are not referenced as substrings in any
// content string. Exported for testing.
func findOrphans(keys []string, contents []string) []string {
	allText := strings.Join(contents, "\n")
	var orphans []string
	for _, key := range keys {
		if !strings.Contains(allText, key) {
			orphans = append(orphans, key)
		}
	}
	return orphans
}
