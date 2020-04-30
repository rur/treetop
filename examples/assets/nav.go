package assets

import (
	"fmt"
	"strings"
)

// Navigation Enum
const (
	NoPage = iota - 1
	IntroNav
	GreeterNav
	InlineNav
	TicketsNav
)

// NavHTML returns the app navigation template for a page given the page numbers
func NavHTML(nav int) string {
	items := []string{
		`<li class="nav-item"><a class="nav-link" href="/" title="Intro">Intro</a></li>`,
		`<li class="nav-item"><a class="nav-link" href="/greeter" title="View Greeter">Greeter</a></li>`,
		`<li class="nav-item"><a class="nav-link" href="/inline" title="Inline Edit">Inline Edit</a></li>`,
		`<li class="nav-item"><a class="nav-link" href="/ticket" title="Ticket Wizard">Ticket Wizard</a></li>`,
		`<li class="nav-item"><a class="nav-link" href="/more" title="More...">More...</a></li>`,
	}

	if nav < NoPage || nav >= len(items) {
		panic(fmt.Sprintf("Invalid page %d", nav))
	}

	if nav != NoPage {
		items[nav] = strings.Replace(items[nav], "nav-link", "nav-link active", 1)
	}
	return fmt.Sprintf(`
	<nav>
		<ul class="nav nav-pills justify-content-center">
			%s
		</ul>
	</nav>
	`, strings.Join(items, "\n"))
}
