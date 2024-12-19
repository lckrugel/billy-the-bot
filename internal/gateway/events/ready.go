package events

type ReadyEvent struct {
	Api_version int    `json:"api_version"`
	Session_id  string `json:"session_id"`
	Resume_url  string `json:"resume_gateway_url"`
	// TODO: User
	// TODO: UnavailableGuilds
	// TODO: Shards
}
