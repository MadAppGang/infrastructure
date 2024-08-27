package main

import (
	"regexp"

	"github.com/charmbracelet/bubbles/viewport"
)

type sesView struct {
	detailViewModel

	s ses
}

func newSesView(e env) *sesView {
	m := &sesView{
		detailViewModel: detailViewModel{
			title:       "AWS Simple Email Service settings",
			description: "Optional AWS Simple Email Service settings",
			inputs: []inputModel{
				newBoolFieldModel(baseInputModel{
					title:       "Enable AWS SES",
					description: "If you want use AWS SES as your email service, enable it.",
				}, boolValue{e.ses.enabled}),
				newTextFieldModel(baseInputModel{
					title:             "Domain name for AWS SES",
					description:       "What email domain name you want to use for AWS SES",
					placeholder:       "email.madappgang.com",
					validator:         regexp.MustCompile(`^(([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,6})?$`),
					validationMessage: "Valid domain name, with no scheme and no query string",
				}, stringValue{e.ses.domainName}),
				newTextFieldModel(baseInputModel{
					title:       "The list of test emails",
					description: "By default your SES will be in sandbox mode, only verified emails could receive emails. You need to write to support team. Before that you can send emails to these email addresses.",
				}, sliceValue{e.ses.testEmails}),
			},
		},
		s: e.ses,
	}

	m.viewport = viewport.New(0, 0)
	m.updateViewportContent()
	return m
}
