package client_announcer

type kafkaMsg struct {
	ChannelID string `json:"channel_id"`
	Username  string `json:"username"`
	Message   string `json:"message"`
	Cluster   string `json:"cluster"`
}

