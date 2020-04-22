package inputs

import (
	"log"
	"net/url"
	"strings"
)

// SoftwareTicket is the model for a helpdesk ticket
type SoftwareTicket struct {
	Summary     string
	Assignees   []string
	IssueType   string
	Description string
	Attachments []*FileInfo
}

// SoftwareTicketFromQuery descodes a help desk ticket model from url query parameters
// Note, unicode normalization omitted for brevity
func SoftwareTicketFromQuery(query url.Values) *SoftwareTicket {
	ticket := &SoftwareTicket{
		Summary:     strings.TrimSpace(query.Get("summary")),
		Assignees:   query["assignees"],
		IssueType:   query.Get("issue-type"),
		Description: strings.TrimSpace(query.Get("description")),
	}

	switch ticket.IssueType {
	case "bug", "enhancement", "epic", "task", "wishlist":
	default:
		ticket.IssueType = ""
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

func (t *SoftwareTicket) RawQuery() string {
	query := url.Values{}
	query.Set("department", "software")
	query.Set("summary", t.Summary)
	query.Set("issue-type", t.IssueType)
	for _, assignee := range t.Assignees {
		query.Add("assignees", assignee)
	}
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
