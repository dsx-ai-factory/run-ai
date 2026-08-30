package utils

import "time"

type CategoryKey string

const (
	NotificationsCategory CategoryKey = "Notifications"
	UIWorkloadsCategory   CategoryKey = "UI-Workloads"
	UIAppearanceCategory  CategoryKey = "UI-Appearance"
	BrandingCategory      CategoryKey = "Branding"
)

func GetCurrentTimestampMs() int64 {
	return time.Now().UnixMilli()
}
