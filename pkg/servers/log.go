package servers

import (
		"log/slog"
)

func logAttributes(log *slog.Logger, missing, extra []string) {
	if len(missing) > 0 {
		log.Info("missing attributes",
			slog.Any("attributes", missing),
		)
	}
	if len(extra) > 0 {
		log.Info("extra attributes",
			slog.Any("attributes", extra),
		)
	}
}
