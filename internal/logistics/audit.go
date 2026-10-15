package logistics

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type Key struct {
	ID     string   `json:"key_id"`
	Name   string   `json:"name"`
	Scopes []string `json:"scopes"`
}

type ShipmentCase struct {
	ShipmentID string          `json:"shipment_id"`
	Events     []ShipmentEvent `json:"events"`
	PODFiles   []PODFile       `json:"pod_files"`
	Exceptions []Exception     `json:"exceptions"`
}

type ShipmentEvent struct {
	Type       string    `json:"type"`
	OccurredAt time.Time `json:"occurred_at"`
}

type PODFile struct {
	Name      string `json:"name"`
	MediaType string `json:"media_type"`
}

type Exception struct {
	Code   string `json:"code"`
	Status string `json:"status"`
}

type KeyReach struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Reaches []string `json:"reaches"`
	Missing []string `json:"missing"`
}

type Review struct {
	ReviewID string       `json:"review_id"`
	Date     string       `json:"date"`
	Shipment ShipmentCase `json:"shipment"`
	Keys     []KeyReach   `json:"keys"`
	PDF      any          `json:"pdf"`
}

var requiredScopes = map[string]string{
	"logistics.shipment_events.read": "shipment events",
	"logistics.pod_files.read":       "proof-of-delivery files",
	"logistics.exceptions.manage":    "exception handling",
	"pdf.generate":                   "dated access-review document",
}

func AuditKeys(keys []Key) []KeyReach {
	result := make([]KeyReach, 0, len(keys))
	for _, key := range keys {
		owned := make(map[string]bool, len(key.Scopes))
		for _, scope := range key.Scopes {
			owned[scope] = true
		}

		reach := KeyReach{ID: key.ID, Name: key.Name, Reaches: []string{}, Missing: []string{}}
		for scope, resource := range requiredScopes {
			if owned[scope] {
				reach.Reaches = append(reach.Reaches, resource)
			} else {
				reach.Missing = append(reach.Missing, resource)
			}
		}
		sort.Strings(reach.Reaches)
		sort.Strings(reach.Missing)
		result = append(result, reach)
	}
	return result
}

func ReviewMarkdown(review Review) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Logistics access review: %s\n\n", review.Date)
	fmt.Fprintf(&b, "Shipment: `%s`\n\n", review.Shipment.ShipmentID)
	fmt.Fprintf(&b, "Observed: %d events, %d proof-of-delivery files, %d exceptions.\n\n", len(review.Shipment.Events), len(review.Shipment.PODFiles), len(review.Shipment.Exceptions))
	for _, key := range review.Keys {
		fmt.Fprintf(&b, "## %s (`%s`)\n\n", key.Name, key.ID)
		fmt.Fprintf(&b, "Reachable: %s.\n\n", strings.Join(key.Reaches, ", "))
		fmt.Fprintf(&b, "Missing: %s.\n\n", strings.Join(key.Missing, ", "))
	}
	return b.String()
}
