package context

import "strings"

type outputBudget struct {
	max     int
	used    int
	omitted []string
}

func newOutputBudget(maxBytes int) *outputBudget {
	if maxBytes <= 0 {
		return nil
	}
	return &outputBudget{max: maxBytes}
}

func (o *outputBudget) take(label, content string) (string, bool) {
	if content == "" {
		return "", true
	}
	if o == nil {
		return content, true
	}
	if o.used+len(content) > o.max {
		o.omitted = append(o.omitted, label)
		return "", false
	}
	o.used += len(content)
	return content, true
}

func (o *outputBudget) omittedNote() string {
	if o == nil || len(o.omitted) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("## Omitted (context --max-bytes)\n\n")
	for _, label := range o.omitted {
		sb.WriteString("- ")
		sb.WriteString(label)
		sb.WriteString("\n")
	}
	return sb.String()
}
