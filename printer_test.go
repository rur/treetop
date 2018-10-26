package treetop

import "testing"

func Test_previewTemplate(t *testing.T) {
	type args struct {
		str    string
		before int
		after  int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "basic",
			args: args{
				str:    "some/path/on/fs.html",
				before: 10,
				after:  10,
			},
			want: "\"some/path/on/fs.html\"",
		},
		{
			name: "realistic html",
			args: args{
				str:    fromExampleDotCom,
				before: 10,
				after:  10,
			},
			want: "\"<!doctype……y></html>\"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := previewTemplate(tt.args.str, tt.args.before, tt.args.after); got != tt.want {
				t.Errorf("previewTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}

var fromExampleDotCom = `
<!doctype html>
<html>
<head>
    <title>Example Domain</title>

    <meta charset="utf-8" />
    <meta http-equiv="Content-type" content="text/html; charset=utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <style type="text/css">
    body {
        background-color: #f0f0f2;
        margin: 0;
        padding: 0;
        font-family: "Open Sans", "Helvetica Neue", Helvetica, Arial, sans-serif;

    }
    div {
        width: 600px;
        margin: 5em auto;
        padding: 50px;
        background-color: #fff;
        border-radius: 1em;
    }
    a:link, a:visited {
        color: #38488f;
        text-decoration: none;
    }
    @media (max-width: 700px) {
        body {
            background-color: #fff;
        }
        div {
            width: auto;
            margin: 0 auto;
            border-radius: 0;
            padding: 1em;
        }
    }
    </style>
</head>

<body>
<div>
    <h1>Example Domain</h1>
    <p>This domain is established to be used for illustrative examples in documents. You may use this
    domain in examples without prior coordination or asking for permission.</p>
    <p><a href="http://www.iana.org/domains/example">More information...</a></p>
</div>
</body>
</html>
`
