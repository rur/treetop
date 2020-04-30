package inputs

import (
	"log"
	"net/url"
	"strings"
)

// SystemsTicket is the model for a helpdesk ticket
type SystemsTicket struct {
	Summary     string
	Description string
	Attachments []*FileInfo
	Components  []string
}

// SystemsTicketFromQuery descodes a help desk ticket model from url query parameters
// Note, unicode normalization omitted for brevity
func SystemsTicketFromQuery(query url.Values) *SystemsTicket {
	ticket := &SystemsTicket{
		Summary:     strings.TrimSpace(query.Get("summary")),
		Description: strings.TrimSpace(query.Get("description")),
	}

	for _, tag := range query["tags"] {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			ticket.Components = append(ticket.Components, tag)
		}
	}

	for _, enc := range query["attachment"] {
		info := &FileInfo{}
		if err := info.UnmarshalBase64([]byte(enc)); err != nil {
			log.Println("Error parsing encoded attachmet", err)
			continue
		}
		ticket.Attachments = append(ticket.Attachments, info)
	}
	return ticket
}

// RawQuery will encode state of systems ticket into a URL encoded query string
func (t *SystemsTicket) RawQuery() string {
	query := url.Values{}
	query.Set("department", "systems")
	query.Set("summary", t.Summary)
	query.Set("description", t.Description)
	for _, att := range t.Attachments {
		enc, err := att.MarshalBase64()
		if err != nil {
			log.Println("Failed to encode attachment", err)
			// skip it
			continue
		}
		query["attachment"] = append(query["attachment"], string(enc))
	}
	for _, tag := range t.Components {
		query.Add("tags", tag)
	}
	return query.Encode()
}
