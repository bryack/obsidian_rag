package domain

import "strings"

type DefaultFormatter struct{}

func (f *DefaultFormatter) Format(doc Document) string {
	var builder strings.Builder

	builder.WriteString("File: ")
	builder.WriteString(doc.FilePath)
	builder.WriteString("\n")

	if len(doc.HeaderPath) > 0 {
		builder.WriteString("Section: ")
		builder.WriteString(strings.Join(doc.HeaderPath, " / "))
		builder.WriteString("\n\n")
	}

	builder.WriteString(doc.Content)
	return builder.String()
}
