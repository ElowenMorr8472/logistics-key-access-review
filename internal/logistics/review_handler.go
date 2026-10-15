package logistics

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type ReviewHandler struct {
	client *Client
	now    func() time.Time
}

func NewReviewHandler(client *Client) http.Handler {
	return &ReviewHandler{client: client, now: time.Now}
}

func (h *ReviewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var shipment ShipmentCase
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&shipment); err != nil || shipment.ShipmentID == "" {
		http.Error(w, "invalid shipment case", http.StatusBadRequest)
		return
	}

	keys, err := h.client.ListKeys(r.Context())
	if err != nil {
		h.writeUpstreamError(w, err)
		return
	}
	review := Review{
		ReviewID: "access-review-" + shipment.ShipmentID + "-" + h.now().UTC().Format("2006-01-02"),
		Date:     h.now().UTC().Format("2006-01-02"),
		Shipment: shipment,
		Keys:     AuditKeys(keys),
	}
	review.PDF, err = h.client.GenerateReview(r.Context(), ReviewMarkdown(review), review.ReviewID)
	if err != nil {
		h.writeUpstreamError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(review)
}

func (h *ReviewHandler) writeUpstreamError(w http.ResponseWriter, err error) {
	var apiErr *InfraiError
	if errors.As(err, &apiErr) && apiErr.Status >= 400 && apiErr.Status < 500 {
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}
	http.Error(w, "upstream request failed", http.StatusBadGateway)
}
