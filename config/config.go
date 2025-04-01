package config

import "os"

var (
	GoogleClientID     = os.Getenv("GOOGLE_CLIENT_ID")
	GoogleClientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
	RedirectURI        = os.Getenv("GOOGLE_REDIRECT_URI")
)
