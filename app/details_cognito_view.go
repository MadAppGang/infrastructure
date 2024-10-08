package main

import (
	"regexp"

	"github.com/charmbracelet/bubbles/viewport"
)

type cognitoView struct {
	detailViewModel

	c Cognito
}

func newCognitoView(e Env) *cognitoView {
	m := &cognitoView{
		detailViewModel: detailViewModel{
			title:       "AWS Cognito settings",
			description: "Optional AWS Cognito settings",
			inputs: []inputModel{
				newBoolFieldModel(baseInputModel{
					title:       "Enable Cognito",
					description: "If you want use cognito as you authentication provider, enable it. But we recommend look to the Aooth first.",
				}, boolValue{e.Cognito.Enabled}),
				newBoolFieldModel(baseInputModel{
					title:       "Cognito Web client enabled",
					description: "Enable cognito web client, which will be used for authentication in web apps",
				}, boolValue{e.Cognito.EnableWebClient}),
				newBoolFieldModel(baseInputModel{
					title:       "Cognito API client enabled",
					description: "Enable API client to authenticate with API calls to Cognito",
				}, boolValue{e.Cognito.EnableDashboardClient}),
				newTextFieldModel(baseInputModel{
					title:       "Allowed Callback URLS",
					description: "Which Callbacks allowed for API client",
				}, sliceValue{e.Cognito.DashboardCallbackURLs}),
				newBoolFieldModel(baseInputModel{
					title:       "Enable User Pool Domain",
					description: "Enable Domain specific for Cognito User Pool",
				}, boolValue{e.Cognito.EnableUserPoolDomain}),
				newTextFieldModel(baseInputModel{
					title:             "User Pool domain prefix",
					description:       "Prefix for the user pool",
					placeholder:       "dev",
					validator:         regexp.MustCompile(`^[a-z][a-zA-Z0-9]{1,}$`),
					validationMessage: "letter, and numbers only, min 1 character",
				}, stringValue{e.Cognito.UserPoolDomainPrefix}),
				newBoolFieldModel(baseInputModel{
					title:       "Backend confirm sign up",
					description: "Backend should confirm sign up",
				}, boolValue{e.Cognito.BackendConfirmSignup}),
				newTextFieldModel(baseInputModel{
					title:       "Auto verified attributes",
					description: "The list of attributes that will be verified by Cognito",
				}, sliceValue{e.Cognito.AutoVerifiedAttributes}),
			},
		},
		c: e.Cognito,
	}

	m.viewport = viewport.New(0, 0)
	m.updateViewportContent()
	return m
}

func (m *cognitoView) env(e Env) Env {
	c := Cognito{}
	c.Enabled = m.inputs[0].value().Bool()
	c.EnableWebClient = m.inputs[1].value().Bool()
	c.EnableDashboardClient = m.inputs[2].value().Bool()
	c.DashboardCallbackURLs = m.inputs[3].value().Slice()
	c.EnableUserPoolDomain = m.inputs[4].value().Bool()
	c.UserPoolDomainPrefix = m.inputs[5].value().String()
	c.BackendConfirmSignup = m.inputs[6].value().Bool()
	c.AutoVerifiedAttributes = m.inputs[7].value().Slice()
	e.Cognito = c
	return e
}
