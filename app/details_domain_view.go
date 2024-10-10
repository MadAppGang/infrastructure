package main

import (
	"regexp"

	"github.com/charmbracelet/bubbles/viewport"
)

type backendDomainView struct {
	detailViewModel

	d Domain
}

func newBackendDomainView(e Env) *backendDomainView {
	m := &backendDomainView{
		detailViewModel: detailViewModel{
			title:       "Domain settings",
			description: "Route 57 Domain management",
			inputs: []inputModel{
				newBoolFieldModel(baseInputModel{
					title:       "Enable domain management",
					description: "By default we use autogenerated domain name, but you can enable domain management for custom domain name",
				}, boolValue{e.Domain.Enabled}),
				newBoolFieldModel(baseInputModel{
					title:       "Create domain zone in Route 53",
					description: "If enabled, we create domain zone in Route 53, if disabled, we use existing domain zone by it's name",
				}, boolValue{e.Domain.CreateDomainZone}),
				newTextFieldModel(baseInputModel{
					title:             "Domain name",
					placeholder:       "example.com",
					description:       "just domain name, no http or https, all subdomains will be created automatically, please refer docs for that.",
					validator:         regexp.MustCompile(`^(([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,12})?$`),
					validationMessage: "Valid domain name, with no scheme and no query string",
				}, stringValue{e.Domain.DomainName}),
			},
		},
		d: e.Domain,
	}

	m.viewport = viewport.New(0, 0)
	m.updateViewportContent()
	return m
}

func (b *backendDomainView) env(e Env) Env {
	d := Domain{}
	d.Enabled = b.inputs[0].value().Bool()
	d.CreateDomainZone = b.inputs[1].value().Bool()
	d.DomainName = b.inputs[2].value().String()
	e.Domain = d
	return e
}
