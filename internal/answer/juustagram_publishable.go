package answer

import (
	"errors"

	"github.com/ggmolly/belfast/internal/logger"
	"github.com/ggmolly/belfast/internal/orm"
	"gorm.io/gorm"
)

func isPublishableJuustagramTemplate(template orm.JuustagramTemplate) (bool, error) {
	// Messages must have resolved text before they can be sent to clients.
	if template.MessagePersist == "" {
		// TODO: Replace with explicit publish flag once Juustagram data is curated.
		return false, nil
	}
	if _, err := orm.GetJuustagramLanguage(template.MessagePersist); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Missing language entries indicate unpublished messages.
			// TODO: Emit a validation report for missing Juustagram language keys.
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func logJuustagramSkip(template orm.JuustagramTemplate, commanderID uint32) {
	logger.WithFields(
		"Juustagram/Range",
		logger.FieldValue("message_id", template.ID),
		logger.FieldValue("commander_id", commanderID),
	).Info("Skipped unpublished juustagram message")
}
