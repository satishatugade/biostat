package models

type ZoomMeetingRequest struct {
	Topic           string              `json:"topic"`
	Agenda          string              `json:"agenda"`
	Type            int                 `json:"type"`
	StartTime       string              `json:"start_time"`
	Duration        int                 `json:"duration"`
	Password        string              `json:"password"`
	DefaultPassword bool                `json:"default_password"`
	PreSchedule     bool                `json:"pre_schedule"`
	Settings        ZoomMeetingSettings `json:"settings"`
}

type ZoomMeetingSettings struct {
	HostVideo               bool          `json:"host_video"`
	ParticipantVideo        bool          `json:"participant_video"`
	JoinBeforeHost          bool          `json:"join_before_host"`
	WaitingRoom             bool          `json:"waiting_room"`
	ApprovalType            int           `json:"approval_type"`
	Audio                   string        `json:"audio"`
	MeetingInvitees         []ZoomInvitee `json:"meeting_invitees"`
	AuthenticationException []ZoomInvitee `json:"authentication_exception"`
}

type ZoomInvitee struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type ZoomMeetingResponse struct {
	JoinURL  string `json:"join_url"`
	StartURL string `json:"start_url"`
	ID       int64  `json:"id"`
}
