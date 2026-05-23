package report

import "testing"

func TestRenderSummaryEndsWithSingleNewline(t *testing.T) {
	summary := RenderSummary(Result{})
	if len(summary) == 0 || summary[len(summary)-1] != '\n' {
		t.Fatalf("expected summary to end with newline")
	}
	if len(summary) >= 2 && summary[len(summary)-2:] == "\n\n" {
		t.Fatalf("expected summary to end with a single newline")
	}
}
