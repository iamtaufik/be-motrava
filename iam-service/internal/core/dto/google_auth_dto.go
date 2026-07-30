package dto

// GoogleMobileAuthRequest is the payload sent by a mobile app after Google Sign-In.
type GoogleMobileAuthRequest struct {
	IDToken string `json:"id_token"`
}
