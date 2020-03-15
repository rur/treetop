package greeter

var (
	LandingHTML = `
	<p id="message"><i>Give me someone to say hello to!</i></p>
	`
	ContentHTML = `
	<div id="content">
		<hr>
		<h3 class="mb-3">Treetop Greeter</h3>
		<form action="/greeter/greet" treetop>
			<div class="input-group mb-3">
				<input id="name"
					name="name"
					type="text"
					autofocus tabindex="0"
					class="form-control"
					aria-label="Name of the person who is to be greeted"
					placeholder="Name of person to greet"
					value="{{ .Value }}">
				<div class="input-group-append">
				<button
					treetop-submitter
					name="submitter"
					value="Greet Me"
					type="button"
					tabindex="1"
					class="btn btn-outline-secondary">Greet Me</button>
				</div>
			</div>
		</form>

		{{ template "message" .Message}}
	</div>
	`
	GreetingHTML = `
	<div id="message" class="mt-4 text-center">
		<h1>Hello, {{ .Who }}!</h1>
		<p><a href="/greeter" treetop>Clear</a></p>

		<div class="alert alert-info small" role="alert">
			{{ .Notes }}
		</div>
		{{ if not .IsFullPage }}
		<div class="alert alert-secondary small" role="alert">
			Browser location and history were updated, try using the back button.
		</div>
		{{ end }}
	</div>
	`
)
