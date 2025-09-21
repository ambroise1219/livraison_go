package delivery

import (
	"github.com/ambroise1219/livraison_go/models"
)

// DeliveryContext représente le contexte d'une livraison pour l'assignation
type DeliveryContext struct {
	Type             models.DeliveryType `json:"type"`
	FirstLat         float64             `json:"firstLat"`
	FirstLng         float64             `json:"firstLng"`
	Stops            int                 `json:"stops"`
	RequiredCapacity int                 `json:"requiredCapacity"`
}

// Note: CourierCandidate est défini dans assignment_service.go

// AssignmentResult représente le résultat d'une assignation
type AssignmentResult struct {
	Success      bool            `json:"success"`
	DriverID     string          `json:"driverId,omitempty"`
	Score        float64         `json:"score,omitempty"`
	ErrorMessage string          `json:"errorMessage,omitempty"`
	DeliveryID   string          `json:"deliveryId"`
	Context      DeliveryContext `json:"context"`
}

// AssignmentStats représente les statistiques d'assignation
type AssignmentStats struct {
	TotalAssignments      int     `json:"totalAssignments"`
	SuccessfulAssignments int     `json:"successfulAssignments"`
	FailedAssignments     int     `json:"failedAssignments"`
	AverageScore          float64 `json:"averageScore"`
	AverageDistance       float64 `json:"averageDistance"`
}

// AssignmentCriteria représente les critères d'assignation
type AssignmentCriteria struct {
	MaxDistance     float64 `json:"maxDistance"`     // Distance maximale en km
	MinScore        float64 `json:"minScore"`        // Score minimum requis
	RequiredVehicle string  `json:"requiredVehicle"` // Type de véhicule requis
	Priority        int     `json:"priority"`        // Priorité (1-10)
}

// AssignmentOptions représente les options d'assignation
type AssignmentOptions struct {
	AutoAssign     bool               `json:"autoAssign"`     // Assignation automatique
	NotifyDriver   bool               `json:"notifyDriver"`   // Notifier le livreur
	RequireConfirm bool               `json:"requireConfirm"` // Requérir confirmation
	Criteria       AssignmentCriteria `json:"criteria"`       // Critères d'assignation
	Timeout        int                `json:"timeout"`        // Timeout en secondes
}
