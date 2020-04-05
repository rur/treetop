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
	.hide{
		position: absolute;
		height: 1px;
		width: 1px;
		overflow: hidden;
		clip: rect(1px, 1px, 1px, 1px);
	}
	.network-inspector {
		position: absolute;
		right: 0px;
		transform: rotate(90deg);
		top: 50vh;
	}
	</style>
</head>
<body>
	<main role="main" class="mt-2">
		<div class="network-inspector d-none d-lg-block text-muted">↑ Expand your Network Inspector ↑</div>
		<div class="container">
		{{ template "nav" .}}

		{{ template "content" . }}
		</div>
	</main>
{{ block "treetop-config" . }}{{ end }}
<script src="/treetop.js" async></script>
</body>
</html>
	`
)
