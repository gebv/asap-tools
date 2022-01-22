package clickup

import "testing"

func Test_teamIDFromURL(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"https://app.clickup.com/2431928/v/f/96471870/42552884", "2431928"},
		{"https://app.clickup.com/2431928/v/li/174386179", "2431928"},
		{"https://app.clickup.com/2431928", "2431928"},
		{"https://app.clickup.com/", ""},
		{"https://app.clickup.com", ""},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := teamIDFromURL(tt.in); got != tt.want {
				t.Errorf("teamIDFromURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_listIDFromURL(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"https://app.clickup.com/2431928/v/li/174318787", "174318787"},
		{"https://app.clickup.com/2431928/v/f/96471870/42552884", ""},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := listIDFromURL(tt.in); got != tt.want {
				t.Errorf("listIDFromURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_folderIDFromURL(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"https://app.clickup.com/2431928/v/li/174318787", ""},
		{"https://app.clickup.com/2431928/v/f/96471870/42552884", "96471870"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := folderIDFromURL(tt.in); got != tt.want {
				t.Errorf("folderIDFromURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
