package services

// SettingsAPIService implements the generated SettingsAPIServicer interface by
// embedding the four per-family handler structs. The OpenAPI tag was
// consolidated to "Settings", so the generator emits a single interface
// covering branding, notifications, workloads, and appearance methods.
type SettingsAPIService struct {
	*BrandingSettingsAPIService
	*NotificationsSettingsAPIService
	*WorkloadsSettingsAPIService
	*AppearanceSettingsAPIService
}

func NewSettingsAPIService() *SettingsAPIService {
	return &SettingsAPIService{
		BrandingSettingsAPIService:      NewBrandingSettingsAPIService(),
		NotificationsSettingsAPIService: NewNotificationsSettingsAPIService(),
		WorkloadsSettingsAPIService:     NewWorkloadsSettingsAPIService(),
		AppearanceSettingsAPIService:    NewAppearanceSettingsAPIService(),
	}
}
