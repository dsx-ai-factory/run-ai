package route

func (brcp *BasicRouteConfigurationProvider) GetNotifyIds() []string {
	return brcp.config.NotifyIds
}
