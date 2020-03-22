package assets

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
	.pointer {
		cursor: pointer;
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
		</div>
		{{ end }}
		</div>
	</main>
{{ block "treetop-config" . }}{{ end }}
<script src="/treetop.js" async></script>
</body>
</html>
	`
)
