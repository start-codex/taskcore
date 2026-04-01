// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package email

import (
	"errors"
	"fmt"
	"log/slog"
)

var ErrSMTPNotConfigured = errors.New("SMTP not configured")

type SMTPConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	From     string `json:"from"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

func (c SMTPConfig) Validate() error {
	if c.Host == "" {
		return errors.New("smtp host is required")
	}
	if c.Port <= 0 {
		return errors.New("smtp port must be positive")
	}
	if c.From == "" {
		return errors.New("smtp from address is required")
	}
	return nil
}

type Message struct {
	To      string
	Subject string
	Body    string // HTML body
}

// Send sends a message using the provided SMTP config.
// If config is nil or empty, logs a warning and returns nil (graceful degradation).
func Send(config *SMTPConfig, msg Message) error {
	if config == nil || config.Host == "" {
		slog.Warn("email not sent: SMTP not configured", "to", msg.To, "subject", msg.Subject)
		return nil
	}
	if err := sendViaSMTP(*config, msg); err != nil {
		return fmt.Errorf("send email to %s: %w", msg.To, err)
	}
	return nil
}
