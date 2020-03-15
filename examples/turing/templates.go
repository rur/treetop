package turing

const ContentHTML = `
	<div id="content">
		<hr>
		<h3 class="mb-3">Turing Test</h3>

		{{ if .ErrorMessage }}
		<div class="alert alert-danger" role="alert">
			<pre ><code>{{ .ErrorMessage }}</code></pre>
		</div>
		{{ end }}

		<h3>JSON</h3>
		<pre ><code>{{ .Data }}</code></pre>

		<h3>Compressed</h3>
		<pre ><code>{{ .Compressed }}</code></pre>

	</div>
	`
