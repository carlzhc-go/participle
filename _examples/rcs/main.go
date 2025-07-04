package main

import (
	"os"
	"github.com/alecthomas/kong"
	"github.com/alecthomas/repr"
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

type Boolean bool

func (b *Boolean) Capture(values []string) error {
	switch values[0] {
	case "", "TRUE", "True", "T", "true", "t", "YES", "Yes", "Y", "yes", "y", "1":
		*b = true
	default:
		*b = false
	}

	return nil
}

type RCS struct {
	Pos lexer.Position
	Admin Admin `parser:"@@"`
	Delta []Delta `parser:"@@*"`
	Desc string `parser:"'desc' @String"`
	Deltatext []Deltatext `parser:"@@*"`
}

// Admin
type Admin struct {
	Pos lexer.Position
	Head float32 `parser:"'head' @Number ';'"`
	Branch string `parser:"( 'branch' @Number ';' )?"`
	Access []string `parser:"'access' @Ident* ';'"`
	Symbols []Symbol `parser:"'symbols' @@* ';'"`
	Locks []Lock `parser:"'locks' @@* ';'"`
	Strict bool `parser:"( @'strict' ';' )?"`
	Integrity string `parser:"( 'integrity' '@' @Intstring '@' ';')?"`
	Comment string `parser:"( 'comment' @String ';' )?"`
	Expand string `parser:"( 'expand' @String ';' )?"`
}

type Symbol struct {
	Pos lexer.Position
	Sym string `parser:"@Ident ':'"`
	Num float32 `parser:"@Number"`
}

type Lock struct {
	Pos lexer.Position
	Id string `parser:"@Ident ':'"`
	Num *float32 `parser:"@Number"`
}


// Delta
type Delta struct {
	Num float32 `parser:"@Number"`
	Date string `parser:"'date' @Number ';'"`
	Author string `parser:"'author' @Ident ';'"`
	State string `parser:"'state' @Ident ';'"`
	Branches []string `parser:"'branches' @Number* ';'"`
	Next float32 `parser:"'next' @Number? ';'"`
	Commitid Symbol `parser:"( 'commentid' @@ ';' )?"`
}

// Deltatext
type Deltatext struct {
	Num float32 `parser:"@Number"`
	Log string `parser:"'log' @String"`
	Text string `parser:"'text' @String"`
}

var (
	rcsLexer = lexer.MustSimple([]lexer.SimpleRule{
		{"Number", `[0-9][.0-9]+`},
		{`Keyword`, `\b(head|branch|access|symbols|locks|integrity|comment|expand|date|author|state|branches|next|commitid|desc|log|text|strict)\b`},
		{"String", `(@[^@]*@)+`},
		{"Intstring", `@[^@]*@`},
		{`Ident`, `[a-zA-Z0-9_~!#%^&*()_+=\]\[{}|\\"<>~.-]+`},
		{"whitespace", `\s+`},
		{"Special", `[$,.:;@]`},
	})

	rcsParser = participle.MustBuild[RCS](
		participle.Lexer(rcsLexer),
	)

	cli struct {
		File string `help:"RCS file to parse." arg:""`
	}
)

func main() {
	ctx := kong.Parse(&cli)
	var r *os.File
	var err error
	r, err = os.Open(cli.File)
	ctx.FatalIfErrorf(err)
	defer r.Close()
	rcs, err := rcsParser.Parse(cli.File, r)
	ctx.FatalIfErrorf(err)
	repr.Println(rcs)
}
