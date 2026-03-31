package bdd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/cucumber/godog"
	"github.com/steveyegge/beads/internal/testutil"
)

func TestMain(m *testing.M) {
	os.Setenv("BEADS_TEST_MODE", "1")
	// Ensure Dolt is available via testcontainers
	if err := testutil.EnsureDoltContainerForTestMain(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting Dolt container: %v\n", err)
		os.Exit(1)
	}
	defer testutil.TerminateDoltContainer()

	code := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
		},
	}.Run()

	if code != 0 {
		os.Exit(code)
	}
}

type testContext struct {
	lastOutput string
	lastError  error
	variables  map[string]string
}

func (c *testContext) runBD(args ...string) error {
	// Add --sandbox to prevent git push/pull during tests
	finalArgs := append([]string{"--sandbox"}, args...)
	cmd := exec.Command("bd", finalArgs...)
	// Ensure we use the test server port
	if port := testutil.DoltContainerPort(); port != "" {
		cmd.Env = append(os.Environ(), "BEADS_DOLT_PORT="+port)
	}
	out, err := cmd.CombinedOutput()
	c.lastOutput = string(out)
	c.lastError = err
	return nil
}

func (c *testContext) aFreshBeadsDatabase() error {
	// Clean up and init
	os.RemoveAll("/root/.beads")
	return c.runBD("init", "--prefix", "bdd", "--quiet")
}

func (c *testContext) iCreateANewTaskWithTitleAndPriority(title string, priority int) error {
	return c.runBD("create", title, "-p", fmt.Sprintf("%d", priority))
}

func (c *testContext) theOutputShouldContain(expected string) error {
	if !strings.Contains(c.lastOutput, expected) {
		return fmt.Errorf("expected output to contain %q, but got:\n%s", expected, c.lastOutput)
	}
	return nil
}

func (c *testContext) theOutputShouldNotContain(expected string) error {
	if strings.Contains(c.lastOutput, expected) {
		return fmt.Errorf("expected output to not contain %q, but it did:\n%s", expected, c.lastOutput)
	}
	return nil
}

func (c *testContext) iListTheIssues() error {
	return c.runBD("list")
}

func (c *testContext) iMarkTheIssueIDAs(varName string) error {
	// Extract ID from output (e.g. "Created issue: bdd-rwv")
	lines := strings.Split(c.lastOutput, "\n")
	for _, line := range lines {
		if strings.Contains(line, "bdd-") {
			parts := strings.Fields(line)
			for _, p := range parts {
				if strings.HasPrefix(p, "bdd-") {
					c.variables[varName] = strings.TrimSuffix(p, ":")
					return nil
				}
			}
		}
	}
	return fmt.Errorf("could not find issue ID in output: %s", c.lastOutput)
}

func (c *testContext) iUpdateTheStatusOfTo(varName, status string) error {
	id := c.variables[varName]
	return c.runBD("update", id, "--status", status)
}

func (c *testContext) iShowTheDetailsOf(varName string) error {
	id := c.variables[varName]
	return c.runBD("show", id)
}

func (c *testContext) iCloseTheIssueWithReason(varName, reason string) error {
	id := c.variables[varName]
	return c.runBD("close", id, "--reason", reason)
}

func (c *testContext) iListTheReadyIssues() error {
	return c.runBD("ready")
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	c := &testContext{
		variables: make(map[string]string),
	}

	ctx.Step(`^a fresh beads database$`, c.aFreshBeadsDatabase)
	ctx.Step(`^I create a new task with title "([^"]*)" and priority (\d+)$`, c.iCreateANewTaskWithTitleAndPriority)
	ctx.Step(`^the output should contain "([^"]*)"$`, c.theOutputShouldContain)
	ctx.Step(`^the output should not contain "([^"]*)"$`, c.theOutputShouldNotContain)
	ctx.Step(`^I list the issues$`, c.iListTheIssues)
	ctx.Step(`^I note the issue ID as "([^"]*)"$`, c.iMarkTheIssueIDAs)
	ctx.Step(`^I update the status of "([^"]*)" to "([^"]*)"$`, c.iUpdateTheStatusOfTo)
	ctx.Step(`^I show the details of "([^"]*)"$`, c.iShowTheDetailsOf)
	ctx.Step(`^I close the issue "([^"]*)" with reason "([^"]*)"$`, c.iCloseTheIssueWithReason)
	ctx.Step(`^I list the ready issues$`, c.iListTheReadyIssues)
}
