package config

type PubSubType string

const (
	HalfYearHours              = 24 * 30 * 6
	HalfYearSeconds            = HalfYearHours * 60 * 60
	PSTRedis        PubSubType = "redis"
	PSTNats         PubSubType = "nats"
	PSTKafka        PubSubType = "kafka"
)

type Pubsub struct {
	Enabled                      *bool      `mapstructure:"pubsub_enabled"`
	Type                         PubSubType `mapstructure:"pubsub_type" validate:"required"`
	Url                          string     `mapstructure:"pubsub_url" validate:"required"`
	User                         string     `mapstructure:"pubsub_user" config:"secret"`
	Password                     string     `mapstructure:"pubsub_password" config:"secret"`
	EnableAdditionalSubscriber   bool       `mapstructure:"pubsub_enable_additional_subscriber"`
	AdditionalSubscriberType     PubSubType `mapstructure:"pubsub_additional_subscriber_type"`
	AdditionalSubscriberUrl      string     `mapstructure:"pubsub_additional_subscriber_url"`
	AdditionalSubscriberUser     string     `mapstructure:"pubsub_additional_subscriber_user" config:"secret"`
	AdditionalSubscriberPassword string     `mapstructure:"pubsub_additional_subscriber_password" config:"secret"`
	EnableTLS                    bool       `mapstructure:"pubsub_enable_tls"`
	SubscribeRetries             int        `mapstructure:"pubsub_subscribe_retries" validate:"gte=0"`
	PublishRetries               int        `mapstructure:"pubsub_publish_retries" validate:"gte=0"`
	ConnectRetries               int        `mapstructure:"pubsub_connect_retries" validate:"gte=0"`
	EnableDLQ                    bool       `mapstructure:"pubsub_enable_dlq"`
	SubscriptionInitRetries      int        `mapstructure:"pubsub_subscription_init_retries" validate:"gte=0"`
	WaitForAckSec                int        `mapstructure:"pubsub_wait_for_ack_sec"`
	PublishRetryInterval         int        `mapstructure:"pubsub_publish_retry_interval_sec"`
	RetryOnReject                bool       `mapstructure:"pubsub_retry_on_reject"`
	ConnectTimeoutSeconds        int        `mapstructure:"pubsub_connect_timeout_seconds"`
	CleanerIntervalSeconds       int        `mapstructure:"pubsub_cleaner_interval_seconds"`
	ConnMaxIdleTimeSec           int64      `mapstructure:"pubsub_conn_max_idle_time_sec"`
	ConsumerTimeoutSec           int64      `mapstructure:"pubsub_consumer_timeout_sec"` // After this timeout an idle consumer with no pending messages will be removed from the consumer group.
	MaxEntries                   int64      `mapstructure:"pubsub_max_entries"`
	MaxMessageAgeMinutes         int64      `mapstructure:"pubsub_max_message_age_minutes"`
	Replicas                     int        `mapstructure:"pubsub_replicas" validate:"gte=1"`
	ServiceName                  string     `mapstructure:"service_name" validate:"required"` // Connection ownership - service name
}

func (ps *Pubsub) SetDefaults() {
	pubsubEnabled := true
	ps.Enabled = &pubsubEnabled
	ps.EnableTLS = true
	ps.SubscribeRetries = 1
	ps.PublishRetries = 3
	ps.ConnectRetries = 3
	ps.SubscriptionInitRetries = 5
	ps.EnableDLQ = true
	ps.WaitForAckSec = 180
	ps.PublishRetryInterval = 1
	ps.RetryOnReject = false
	ps.ConnectTimeoutSeconds = 2
	ps.CleanerIntervalSeconds = 60
	ps.ConnMaxIdleTimeSec = 5 * 60
	ps.ConsumerTimeoutSec = HalfYearSeconds
	ps.MaxEntries = 500_000
	ps.MaxMessageAgeMinutes = 60 * 24 * 7 // 7 days
	ps.Replicas = 3
	ps.EnableAdditionalSubscriber = false
}
