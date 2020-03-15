package assets

import (
	"fmt"
	"strings"
)

var (
	BaseHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>Treetop Examples</title>
	<link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.4.1/css/bootstrap.min.css">
	<style>
	.container {
		width: auto;
		max-width: 680px;
		padding: 0 15px;
	}
	nav {
		padding: 1rem;
	}
	</style>
</head>
<body>
	<main role="main" class="mt-2">
		<div class="container">
		{{ template "nav" .}}

		{{ block "content" . }}
		<div id="content">
			<p class="text-center">↑ Choose a demo ↑</p>

			<h1 class="mt-5">Treetop Examples</h1>

			<p class="lead">
				These examples demonstrate how full-page and template requests
				can be combined to achieve various interactive controls.
			</p>

			<h3 class="mt-4">Writer vs. View</h3>

			<p class="mb5">
				Most examples are implemented using hierarchical view.
				However, examples with the 'Writer' label use vanilla HTTP handlers
				for the sake of comparison.
			</p>
		</div>
		{{ end }}
		</div>
	</main>
<script>TREETOP_CONFIG={/*defaults*/}</script>
<script src="/treetop.js" async></script>
</body>
</html>
	`
)

// Navigation Enum
const (
	NoPage = iota - 1
	HomeNav
	GreeterNav
	WriterNav
	TuringNav
	TicketsNav
)

// NavHTML returns the app navigation template for a page given the page numbers
func NavHTML(nav int) string {
	items := []string{
		`<li class="nav-item"><a class="nav-link" href="/" title="Home">Home</a></li>`,
		`<li class="nav-item"><a class="nav-link" href="/greeter" title="View Greeter">Greeter</a></li>`,
		`<li class="nav-item"><a class="nav-link" href="/writer" title="Writer Greeter">Greeter <sup>(Writer)</sup></a></li>`,
		`<li class="nav-item"><a class="nav-link" href="/turing" title="Turing Chat">Turing Chat</a></li>`,
		`<li class="nav-item"><a class="nav-link" href="/ticket" title="Create Ticket">Ticket Form</a></li>`,
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
