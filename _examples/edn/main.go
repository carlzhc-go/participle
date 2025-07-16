package main

import (
	"github.com/alecthomas/kong"
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/alecthomas/repr"
	"os"
)

type Boolean bool

func (b *Boolean) Capture(values []string) error {
	switch values[0] {
	case "true":
		*b = true
	default:
		*b = false
	}

	return nil
}

type Nil bool

func (n *Nil) Capture(values []string) error {
	switch values[0] {
	case "nil":
		*n = true
	default:
		*n = false
	}

	return nil
}

type EDN struct {
	Pos   lexer.Position
	Forms []Form `parser:"@@*"`
}

type Form struct {
	Pos     lexer.Position
	Symbol  *string    `parser:"  @Symbol"`
	Keyword *string    `parser:"| @Keyword"`
	Integer *int64     `parser:"| @Integer"`
	Float   *float64   `parser:"| @Float"`
	List    []Form     `parser:"| '(' @@* ')'"`
	Vector  []Form     `parser:"| '[' @@* ']'"`
	NsMap   NsMap      `parser:"| '#:' @@"`
	Set     []Form     `parser:"| '#{' @@* '}'"`
	Map     []MapEntry `parser:"| '{' @@* '}'"`
	String  *string    `parser:"| @String"`
	Nil     Nil        `parser:"| @Nil"`
}

type NsMap struct {
	Pos   lexer.Position
	Namespace *string `parser:"@Namespace"`
	Map []MapEntry `parser:"'{' @@* '}'"`
}

type MapEntry struct {
	Pos   lexer.Position
	Key   Form `parser:"@@"`
	Value Form `parser:"@@"`
}

var (
	ednLexer = lexer.MustSimple([]lexer.SimpleRule{
		{Name: "Float", Pattern: `[+-]?[0-9]*\.[0-9]+`}, // 64-bit
		{Name: "Integer", Pattern: `[+-]?[0-9]+`},       // 64-bit
		{Name: "Keyword", Pattern: `:[#.*+!_?$%&=<-]?[a-zA-Z0-9][a-zA-Z0-9#.*+!_?$%&=<-]*/?[a-zA-Z0-9#.*+!_?$%&=<-]*`},
		{Name: "String", Pattern: `"(\\"|[^"])*"`},
		{Name: "Character", Pattern: `\b(\\\w|\\u\d{Name: 4})\b`},
		{Name: "Namespace", Pattern: `[A-Za-z0-9_.-]+`},
		{Name: "Symbol", Pattern: `[.+-]?[A-Za-z][a-zA-Z0-9:#.*+!_?$%&=<-]*/?[a-zA-Z0-9:#.*+!_?$%&=<-]+`},
		{Name: "Nil", Pattern: `\bnil\b`},
		{Name: "comment", Pattern: `;[^\n]*|#_(#?\{[^}]*}|\[[^]]*\]|\([^\)]*\)|[^\s]*)`}, // comment
		{Name: "special", Pattern: `#?[][}{)(:]`},
		{Name: "whitespace", Pattern: `[,\s]+`},
})

	ednParser = participle.MustBuild[EDN](
		participle.Lexer(ednLexer),
		participle.Unquote("String"),
		participle.UseLookahead(2),
	)

	cli struct {
		File string `help:"EDN file to parse." arg:""`
	}
)

func main() {
	ctx := kong.Parse(&cli)
	var r *os.File
	var err error
	r, err = os.Open(cli.File)
	ctx.FatalIfErrorf(err)
	defer r.Close()
	edn, err := ednParser.Parse(cli.File, r)
	ctx.FatalIfErrorf(err)
	repr.Println(edn)
}
