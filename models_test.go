package sdk_test

import (
	"encoding/json"
	"testing"

	sdk "github.com/DouDOU-start/airgate-sdk"
)

func TestConfigFieldJSONRoundTrip(t *testing.T) {
	cf := sdk.ConfigField{
		Key:         "api_base",
		Label:       "API Base URL",
		Type:        "string",
		Required:    true,
		Default:     "https://api.example.com",
		Description: "Base URL for API",
		Placeholder: "https://...",
	}

	data, err := json.Marshal(cf)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got sdk.ConfigField
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if got != cf {
		t.Errorf("round-trip mismatch:\ngot  %+v\nwant %+v", got, cf)
	}
}

func TestConfigFieldJSONTags(t *testing.T) {
	cf := sdk.ConfigField{
		Key:      "db_dsn",
		Label:    "Database DSN",
		Type:     "password",
		Required: true,
		// Default, Description, Placeholder left empty (omitempty)
	}

	data, err := json.Marshal(cf)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Unmarshal to map: %v", err)
	}

	// Required keys must be present
	for _, key := range []string{"key", "label", "type", "required"} {
		if _, ok := m[key]; !ok {
			t.Errorf("expected JSON key %q to be present", key)
		}
	}

	// omitempty fields with zero values should be absent
	for _, key := range []string{"default", "description", "placeholder"} {
		if _, ok := m[key]; ok {
			t.Errorf("expected JSON key %q to be omitted for zero value", key)
		}
	}
}

func TestAccountTypeWithFields(t *testing.T) {
	at := sdk.AccountType{
		Key:         "oauth",
		Label:       "OAuth Token",
		Description: "Use OAuth for authentication",
		Fields: []sdk.CredentialField{
			{
				Key:         "access_token",
				Label:       "Access Token",
				Type:        "password",
				Required:    true,
				Placeholder: "sk-...",
			},
			{
				Key:      "refresh_token",
				Label:    "Refresh Token",
				Type:     "password",
				Required: false,
			},
		},
	}

	data, err := json.Marshal(at)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got sdk.AccountType
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if got.Key != at.Key {
		t.Errorf("Key = %q, want %q", got.Key, at.Key)
	}
	if got.Label != at.Label {
		t.Errorf("Label = %q, want %q", got.Label, at.Label)
	}
	if len(got.Fields) != 2 {
		t.Fatalf("Fields length = %d, want 2", len(got.Fields))
	}
	if got.Fields[0].Key != "access_token" {
		t.Errorf("Fields[0].Key = %q, want %q", got.Fields[0].Key, "access_token")
	}
	if got.Fields[1].Required != false {
		t.Errorf("Fields[1].Required = %v, want false", got.Fields[1].Required)
	}
}

func TestCredentialFieldJSONRoundTrip(t *testing.T) {
	cf := sdk.CredentialField{
		Key:         "api_key",
		Label:       "API Key",
		Type:        "password",
		Required:    true,
		Placeholder: "Enter your key",
	}

	data, err := json.Marshal(cf)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got sdk.CredentialField
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if got != cf {
		t.Errorf("round-trip mismatch:\ngot  %+v\nwant %+v", got, cf)
	}
}

func TestCredentialFieldJSONKeys(t *testing.T) {
	cf := sdk.CredentialField{
		Key:   "token",
		Label: "Token",
		Type:  "text",
	}

	data, err := json.Marshal(cf)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Unmarshal to map: %v", err)
	}

	expectedKeys := []string{"key", "label", "type", "required", "placeholder"}
	for _, k := range expectedKeys {
		if _, ok := m[k]; !ok {
			t.Errorf("expected JSON key %q to be present", k)
		}
	}
}

func TestQuotaInfoExtraMapNil(t *testing.T) {
	qi := sdk.QuotaInfo{
		Total:     100.0,
		Used:      25.5,
		Remaining: 74.5,
		Currency:  "USD",
	}

	if qi.Extra != nil {
		t.Errorf("Extra should be nil when not initialized, got %v", qi.Extra)
	}

	// JSON round-trip with nil Extra
	data, err := json.Marshal(qi)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got sdk.QuotaInfo
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if got.Total != qi.Total {
		t.Errorf("Total = %v, want %v", got.Total, qi.Total)
	}
	if got.Currency != qi.Currency {
		t.Errorf("Currency = %q, want %q", got.Currency, qi.Currency)
	}
}

func TestQuotaInfoExtraMapPopulated(t *testing.T) {
	qi := sdk.QuotaInfo{
		Total:     1000.0,
		Used:      200.0,
		Remaining: 800.0,
		Currency:  "CNY",
		ExpiresAt: "2026-12-31T23:59:59Z",
		Extra: map[string]string{
			"plan":   "enterprise",
			"region": "us-east-1",
			"tier":   "premium",
		},
	}

	data, err := json.Marshal(qi)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got sdk.QuotaInfo
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if len(got.Extra) != 3 {
		t.Fatalf("Extra length = %d, want 3", len(got.Extra))
	}

	for k, want := range qi.Extra {
		if gotVal, ok := got.Extra[k]; !ok {
			t.Errorf("Extra missing key %q", k)
		} else if gotVal != want {
			t.Errorf("Extra[%q] = %q, want %q", k, gotVal, want)
		}
	}
}

func TestQuotaInfoExtraMapMutation(t *testing.T) {
	qi := sdk.QuotaInfo{
		Extra: make(map[string]string),
	}

	qi.Extra["key1"] = "val1"
	qi.Extra["key2"] = "val2"

	if len(qi.Extra) != 2 {
		t.Fatalf("Extra length = %d, want 2", len(qi.Extra))
	}

	delete(qi.Extra, "key1")
	if len(qi.Extra) != 1 {
		t.Fatalf("Extra length after delete = %d, want 1", len(qi.Extra))
	}

	if _, ok := qi.Extra["key1"]; ok {
		t.Error("key1 should have been deleted")
	}
	if v := qi.Extra["key2"]; v != "val2" {
		t.Errorf("Extra[key2] = %q, want %q", v, "val2")
	}
}

func TestQuotaInfoJSONWithNullExtra(t *testing.T) {
	// Simulate JSON with explicit null for extra
	raw := `{"total":50,"used":10,"remaining":40,"currency":"EUR","expires_at":"","extra":null}`

	var qi sdk.QuotaInfo
	if err := json.Unmarshal([]byte(raw), &qi); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if qi.Total != 50 {
		t.Errorf("Total = %v, want 50", qi.Total)
	}
	if qi.Extra != nil {
		t.Errorf("Extra should be nil for JSON null, got %v", qi.Extra)
	}
}
