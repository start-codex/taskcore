// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package email

import (
	"fmt"
	"net/smtp"
	"strings"
)

func sendViaSMTP(config SMTPConfig, msg Message) error {
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	headers := []string{
		fmt.Sprintf("From: %s", config.From),
		fmt.Sprintf("To: %s", msg.To),
		fmt.Sprintf("Subject: %s", msg.Subject),
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
	}

	body := strings.Join(headers, "\r\n") + "\r\n\r\n" + msg.Body

	var auth smtp.Auth
	if config.Username != "" && config.Password != "" {
		auth = smtp.PlainAuth("", config.Username, config.Password, config.Host)
	}

	err := smtp.SendMail(addr, auth, config.From, []string{msg.To}, []byte(body))
	if err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	return nil
}
