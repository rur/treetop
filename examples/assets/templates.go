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

func NavHTML(title string) string {
	items := []string{
		`<li class="nav-item"><a class="nav-link" href="/" title="Home">Home</a></li>`,
		`<li class="nav-item"><a class="nav-link" href="/view" title="View Greeter">Greeter (View)</a></li>`,
		`<li class="nav-item"><a class="nav-link" href="/writer" title="Writer Greeter">Greeter (Writer)</a></li>`,
		`<li class="nav-item"><a class="nav-link disabled" href="/" >More... (TODO)</a></li>`,
	}
	switch title {
	case "Home":
		items[0] = strings.Replace(items[0], "nav-link", "nav-link active", 1)
	case "View":
		items[1] = strings.Replace(items[1], "nav-link", "nav-link active", 1)
	case "Writer":
		items[2] = strings.Replace(items[2], "nav-link", "nav-link active", 1)
	}
	return fmt.Sprintf(`
	<nav>
		<ul class="nav justify-content-center">
			%s
		</ul>
	</nav>
	`, strings.Join(items, "\n"))
}
