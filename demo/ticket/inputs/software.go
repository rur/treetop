package inputs

import (
	"log"
	"net/url"
	"strings"
)

// SoftwareTicket is the model for a helpdesk ticket
type SoftwareTicket struct {
	Summary     string
	Assignees   []Assignee
	IssueType   string
	Description string
	Attachments []*FileInfo
}

type Assignee struct {
	Name string
	Role string
}

// SoftwareTicketFromQuery descodes a help desk ticket model from url query parameters
// Note, unicode normalization omitted for brevity
func SoftwareTicketFromQuery(query url.Values) *SoftwareTicket {
	ticket := &SoftwareTicket{
		Summary:     strings.TrimSpace(query.Get("summary")),
		IssueType:   query.Get("issue-type"),
		Description: strings.TrimSpace(query.Get("description")),
	}

	var offset int
	roles := query["assignee-role"]
	for _, name := range query["assignees"] {
		name = strings.TrimSpace(name)
		assignee := Assignee{
			Name: name,
		}
		if name != "" {
			if offset < len(roles) {
				assignee.Role = roles[offset]
				offset++
			}
			ticket.Assignees = append(ticket.Assignees, assignee)
		}
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
		query.Add("assignees", assignee.Name)
		query.Add("assignee-role", assignee.Role)
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
