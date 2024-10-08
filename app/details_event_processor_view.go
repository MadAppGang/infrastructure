package main

import (
	"regexp"

	"github.com/charmbracelet/bubbles/viewport"
)

type eventProcessorTaskView struct {
	detailViewModel

	t EventProcessorTask
}

func NewEventProcessorTaskView(t EventProcessorTask) *eventProcessorTaskView {
	m := &eventProcessorTaskView{
		detailViewModel: detailViewModel{
			title:       "Scheduled ECS task",
			description: "ECR Repository will be crated for the service",
			inputs: []inputModel{
				newTextFieldModel(baseInputModel{
					title:             "Event processor task name",
					description:       "The ECS task which will be run on specific event in event bus.",
					placeholder:       "on_new_order",
					validator:         regexp.MustCompile(`^($|[a-zA-Z][\w-]{3,254})$`),
					validationMessage: "Valid ECS service name, letter, numbers and dash only, min 3 and max 255 characters",
				}, stringValue{t.Name}),
				newTextFieldModel(baseInputModel{
					title:             "Rule name",
					description:       "The Cloudwatch event rule name which will be triggered by the event.",
					placeholder:       "new_order_rule",
					validator:         regexp.MustCompile(`^$|^[a-zA-Z0-9][-_.a-zA-Z0-9]{3,63}$`),
					validationMessage: "Valid Cloudwatch event rule name, letter, numbers and dash only, min 3 and max 63 characters",
				}, stringValue{t.RuleName}),
				newTextFieldModel(baseInputModel{
					title:       "Detail types",
					description: "Optional filter by details types.",
				}, sliceValue{t.DetailTypes}),
				newTextFieldModel(baseInputModel{
					title:       "Sources to catch",
					description: "Optional filter by sources of messages.",
				}, sliceValue{t.Sources}),
				newTextFieldModel(baseInputModel{
					title:       "External Docker image",
					placeholder: "madappgang/aooth",
					description: "Optional Docker hub image name, by default we use private ECR registry for task",
				}, stringValue{t.ExternalDockerImage}),
				newTextFieldModel(baseInputModel{
					title:             "custom docker container command",
					placeholder:       "[\"aooth\", \"--flag\"]",
					description:       "Optional provide container command",
					validator:         regexp.MustCompile(`(^\s*$|\[(\s*"[^"]*"\s*,?\s*)*\])`),
					validationMessage: "Container command is JSON  array of strings format",
				}, stringValue{t.ContainerCommand}),
				newBoolFieldModel(baseInputModel{
					title:       "Allow public access",
					description: "Allow your task to access public internet, it you are using docker image from dockerhub, this should be true.",
				}, boolValue{t.AllowPublicAccess}),
			},
		},
		t: t,
	}

	m.viewport = viewport.New(0, 0)
	m.updateViewportContent()
	return m
}

func (m *eventProcessorTaskView) env(e Env) Env {
	t := EventProcessorTask{}
	t.Name = m.inputs[0].value().String()
	t.RuleName = m.inputs[1].value().String()
	t.DetailTypes = m.inputs[2].value().Slice()
	t.Sources = m.inputs[3].value().Slice()
	t.ExternalDockerImage = m.inputs[4].value().String()
	t.ContainerCommand = m.inputs[5].value().String()
	t.AllowPublicAccess = m.inputs[6].value().Bool()
	e.EventProcessorTasks = append(e.EventProcessorTasks, t)
	return e
}
