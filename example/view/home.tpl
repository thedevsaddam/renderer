{{define "title"}}Home{{end}}
{{define "content"}}
    <h3>Home page</h3>
    <ul>
        <li><a href="/">Home</a></li>
        <li><a href="/about">About Me</a></li>
    </ul>
    <h1>Hello, {{ .Name }}</h1>
    <p>Lorem ipsum dolor sit amet</p>
{{end}}