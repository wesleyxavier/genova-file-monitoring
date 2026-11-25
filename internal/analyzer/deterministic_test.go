package analyzer

import "testing"

func TestRunDeterministic(t *testing.T) {
	code := `package main
<<<<<<< HEAD
const apiKey = "super-secret-value-12345"
func main() {
	fmt.Println("debug")
}`
	findings := RunDeterministic("main.go", code)
	if len(findings) != 3 {
		t.Fatalf("esperava 3 achados determinísticos, obteve %d: %+v", len(findings), findings)
	}
	if findings[0].RuleID != "merge-conflict-marker" || findings[1].RuleID != "hardcoded-secret" || findings[2].RuleID != "no-debug-log" {
		t.Fatalf("regras inesperadas: %+v", findings)
	}
}
