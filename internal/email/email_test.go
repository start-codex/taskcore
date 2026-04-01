// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package email

import (
	"testing"
)

func TestSMTPConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  SMTPConfig
		wantErr bool
	}{
		{name: "valid", config: SMTPConfig{Host: "smtp.example.com", Port: 587, From: "a@b.com"}},
		{name: "missing host", config: SMTPConfig{Port: 587, From: "a@b.com"}, wantErr: true},
		{name: "missing port", config: SMTPConfig{Host: "smtp.example.com", From: "a@b.com"}, wantErr: true},
		{name: "missing from", config: SMTPConfig{Host: "smtp.example.com", Port: 587}, wantErr: true},
		{name: "negative port", config: SMTPConfig{Host: "smtp.example.com", Port: -1, From: "a@b.com"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSend_NilConfig(t *testing.T) {
	msg := Message{To: "test@example.com", Subject: "Test", Body: "<p>Test</p>"}
	// Should not error — graceful degradation
	if err := Send(nil, msg); err != nil {
		t.Fatalf("Send(nil) error = %v, want nil", err)
	}
}

func TestSend_EmptyHost(t *testing.T) {
	config := &SMTPConfig{Host: "", Port: 587, From: "a@b.com"}
	msg := Message{To: "test@example.com", Subject: "Test", Body: "<p>Test</p>"}
	if err := Send(config, msg); err != nil {
		t.Fatalf("Send(empty host) error = %v, want nil", err)
	}
}

func TestRenderTemplate_ValidName(t *testing.T) {
	// password_reset without .html suffix should work
	body, err := RenderTemplate("password_reset", struct{ ResetURL string }{"https://example.com/reset?token=abc"})
	if err != nil {
		t.Fatalf("RenderTemplate error = %v", err)
	}
	if body == "" {
		t.Fatal("RenderTemplate returned empty body")
	}
}

func TestRenderTemplate_WithSuffix(t *testing.T) {
	body, err := RenderTemplate("password_reset.html", struct{ ResetURL string }{"https://example.com/reset"})
	if err != nil {
		t.Fatalf("RenderTemplate error = %v", err)
	}
	if body == "" {
		t.Fatal("RenderTemplate returned empty body")
	}
}

func TestRenderTemplate_Unknown(t *testing.T) {
	_, err := RenderTemplate("nonexistent", nil)
	if err == nil {
		t.Fatal("RenderTemplate should fail for unknown template")
	}
}
