module cz0.cz/czoczo/dotman

go 1.12

require (
	cz0.cz/czoczo/dotman/views v0.0.0
	github.com/namsral/flag v1.7.4-pre
	golang.org/x/crypto v0.0.0-20190820162420-60c769a6c586
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v2 v2.2.2
)

replace cz0.cz/czoczo/dotman/views => ./views
