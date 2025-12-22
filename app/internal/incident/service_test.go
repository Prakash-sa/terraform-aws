package incident

import (
	"context"
	"testing"

	"github.com/Prakash-sa/terraform-aws/app/pkg/models"
)

func TestCreateIncident(t *testing.T) {
	service := NewService(nil)

	req := models.CreateIncidentRequest{
		Title:       "Test Incident",
		Description: "This is a test incident",
		Severity:    models.SeverityHigh,
		Source:      "test",
	}

	incident, err := service.CreateIncident(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to create incident: %v", err)
	}

	if incident.ID == "" {
		t.Error("Expected incident ID to be set")
	}

	if incident.Title != req.Title {
		t.Errorf("Expected title %s, got %s", req.Title, incident.Title)
	}

	if incident.Status != models.StatusOpen {
		t.Errorf("Expected status %s, got %s", models.StatusOpen, incident.Status)
	}
}

func TestGetIncident(t *testing.T) {
	service := NewService(nil)

	// Create an incident first
	req := models.CreateIncidentRequest{
		Title:       "Test Incident",
		Description: "This is a test incident",
		Severity:    models.SeverityMedium,
		Source:      "test",
	}

	created, err := service.CreateIncident(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to create incident: %v", err)
	}

	// Get the incident
	retrieved, err := service.GetIncident(created.ID)
	if err != nil {
		t.Fatalf("Failed to get incident: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("Expected ID %s, got %s", created.ID, retrieved.ID)
	}

	// Try to get a non-existent incident
	_, err = service.GetIncident("non-existent-id")
	if err == nil {
		t.Error("Expected error for non-existent incident, got nil")
	}
}

func TestListIncidents(t *testing.T) {
	service := NewService(nil)

	// Initially, list should be empty
	incidents := service.ListIncidents()
	if len(incidents) != 0 {
		t.Errorf("Expected 0 incidents, got %d", len(incidents))
	}

	// Create a few incidents
	for i := 0; i < 3; i++ {
		req := models.CreateIncidentRequest{
			Title:       "Test Incident",
			Description: "This is a test incident",
			Severity:    models.SeverityLow,
			Source:      "test",
		}
		_, err := service.CreateIncident(context.Background(), req)
		if err != nil {
			t.Fatalf("Failed to create incident: %v", err)
		}
	}

	// List should now have 3 incidents
	incidents = service.ListIncidents()
	if len(incidents) != 3 {
		t.Errorf("Expected 3 incidents, got %d", len(incidents))
	}
}

func TestUpdateIncidentStatus(t *testing.T) {
	service := NewService(nil)

	// Create an incident
	req := models.CreateIncidentRequest{
		Title:       "Test Incident",
		Description: "This is a test incident",
		Severity:    models.SeverityHigh,
		Source:      "test",
	}

	incident, err := service.CreateIncident(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to create incident: %v", err)
	}

	// Update status
	err = service.UpdateIncidentStatus(incident.ID, models.StatusResolved)
	if err != nil {
		t.Fatalf("Failed to update status: %v", err)
	}

	// Verify status was updated
	updated, err := service.GetIncident(incident.ID)
	if err != nil {
		t.Fatalf("Failed to get incident: %v", err)
	}

	if updated.Status != models.StatusResolved {
		t.Errorf("Expected status %s, got %s", models.StatusResolved, updated.Status)
	}

	if updated.ResolvedAt == nil {
		t.Error("Expected ResolvedAt to be set")
	}
}
