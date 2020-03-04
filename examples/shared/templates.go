package shared

// BaseTemplate is the document markup shared by all of the example applications
var BaseTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>Treetop Examples</title>
</head>
<body>

	<nav>
		<ul>
			<li><a href="/view" title="View Greeter">View Greeter</a></li>
			<li><a href="/writer" title="Writer Greeter">Writer Greeter</a></li>
		</ul>
	</nav>

	{{ block "content" . }}
	<p id="content">↑ Choose a demo ↑</p>
	{{ end }}

<script>TREETOP_CONFIG={/*defaults*/}</script>
<script src="/treetop.js" async></script>
</body>
</html>
	`
