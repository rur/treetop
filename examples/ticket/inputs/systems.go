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
}

// SystemsTicketFromQuery descodes a help desk ticket model from url query parameters
// Note, unicode normalization omitted for brevity
func SystemsTicketFromQuery(query url.Values) *SystemsTicket {
	ticket := &SystemsTicket{
		Summary:     strings.TrimSpace(query.Get("summary")),
		Description: strings.TrimSpace(query.Get("description")),
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
	return query.Encode()
}
