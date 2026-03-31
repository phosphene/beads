Feature: Issue Management
  As a developer
  I want to track my project tasks
  So that I can stay organized

  Background:
    Given a fresh beads database

  Scenario: Create and list issues
    When I create a new task with title "E2E BDD Task" and priority 1
    Then the output should contain "E2E BDD Task"
    And the output should contain "bdd-"
    When I list the issues
    Then the output should contain "E2E BDD Task"

  Scenario: Update issue status
    When I create a new task with title "Task to update" and priority 2
    Then the output should contain "bdd-"
    And I note the issue ID as "TASK_ID"
    When I update the status of "TASK_ID" to "in_progress"
    When I show the details of "TASK_ID"
    Then the output should contain "IN_PROGRESS"

  Scenario: Delete/Close an issue
    When I create a new task with title "Task to close" and priority 3
    And I note the issue ID as "TASK_ID"
    When I close the issue "TASK_ID" with reason "Done"
    Then the output should contain "Closed"
    When I list the ready issues
    Then the output should not contain "Task to close"
