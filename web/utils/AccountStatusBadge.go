package utils

import (
	"fmt"
	"html/template"
	"time"
)

func AccountStatusBadge(liftTimestamp *time.Time, isPermanent *bool) template.HTML {
	if isPermanent != nil && *isPermanent {
		return template.HTML(`<span class="badge badge-outline badge-error">Permanently Banned</span>`)
	}
	if liftTimestamp != nil && time.Now().Before(*liftTimestamp) {
		// format timeleft in dd:hh:mm:ss
		timeLeft := time.Until(*liftTimestamp)
		days := int(timeLeft.Hours() / 24)
		hours := int(timeLeft.Hours()) - (days * 24)
		minutes := int(timeLeft.Minutes()) - (days * 24 * 60) - (hours * 60)
		seconds := int(timeLeft.Seconds()) - (days * 24 * 60 * 60) - (hours * 60 * 60) - (minutes * 60)

		return template.HTML(`<span class="badge badge-outline badge-warning" title="Lifts in ` + fmt.Sprintf("%02d", days) + ` days, ` + fmt.Sprintf("%02d", hours) + ` hours, ` + fmt.Sprintf("%02d", minutes) + ` minutes, ` + fmt.Sprintf("%02d", seconds) + ` seconds">Temporarily Banned</span>`)
	}
	return template.HTML(`<span class="badge badge-outline badge-success">Active</span>`)
}
