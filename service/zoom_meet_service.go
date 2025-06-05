package service

import (
	"biostat/models"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

func GetRefreshedZoomAccessToken(refreshToken string) (map[string]interface{}, error) {
	apiUrl := fmt.Sprintf("https://zoom.us/oauth/token?grant_type=refresh_token&refresh_token=%s", url.QueryEscape(refreshToken))
	clientID := os.Getenv("ZOOM_CLIENT_ID")
	clientSecret := os.Getenv("ZOOM_CLIENT_SECRET")

	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", clientID, clientSecret)))

	req, err := http.NewRequest("POST", apiUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Basic "+auth)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		log.Println("Error response:", string(body))
		return nil, errors.New("Failed to get refresh token: " + result["reason"].(string))
	}
	return result, nil
}

func CreateZoomMeeting(accessToken, topic, agenda string, start time.Time, duration int, rawInvitees []map[string]string) (*models.ZoomMeetingResponse, error) {
	invitees := []models.ZoomInvitee{}
	authExceptions := []models.ZoomInvitee{}
	for _, r := range rawInvitees {
		email := r["email"]
		name := r["name"]
		invitees = append(invitees, models.ZoomInvitee{
			Email: email,
		})

		authExceptions = append(authExceptions, models.ZoomInvitee{
			Email: email,
			Name:  name,
		})
	}

	payload := models.ZoomMeetingRequest{
		Topic:           topic,
		Agenda:          agenda,
		Type:            2,
		StartTime:       start.Format("2006-01-02T15:04:05Z"),
		Duration:        duration,
		Password:        "123456",
		DefaultPassword: false,
		PreSchedule:     true,
		Settings: models.ZoomMeetingSettings{
			HostVideo:               true,
			ParticipantVideo:        false,
			JoinBeforeHost:          true,
			WaitingRoom:             false,
			ApprovalType:            2,
			Audio:                   "telephony",
			// MeetingInvitees:         invitees,
			// AuthenticationException: authExceptions,
		},
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	req, err := http.NewRequest("POST", "https://api.zoom.us/v2/users/me/meetings", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Zoom API: %w", err)
	}
	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)
	if res.StatusCode != 201 {
		fmt.Println(res.StatusCode)
		return nil, fmt.Errorf("Zoom API error: %s", string(body))
	}

	var meetingResp models.ZoomMeetingResponse
	if err := json.Unmarshal(body, &meetingResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &meetingResp, nil
}
