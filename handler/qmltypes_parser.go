package handler

import (
	"os"
	"strconv"
	"strings"
)

// ParseQMLTypesFile reads and parses a .qmltypes file from disk.
func ParseQMLTypesFile(path string) (*QMLTypesModule, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseQMLTypes(string(data))
}

// ParseQMLTypes parses the content of a .qmltypes file.
func ParseQMLTypes(content string) (*QMLTypesModule, error) {
	p := &qmltypesParser{src: content}
	return p.parse()
}

// token types for the qmltypes lexer.
type tokKind int

const (
	tokEOF tokKind = iota
	tokIdent
	tokString
	tokNumber
	tokLBrace
	tokRBrace
	tokLBracket
	tokRBracket
	tokColon
	tokComma
	tokSemicolon
	tokTrue
	tokFalse
)

type token struct {
	kind tokKind
	text string
}

type qmltypesParser struct {
	src string
	pos int
	cur token
}

func (p *qmltypesParser) parse() (*QMLTypesModule, error) {
	p.next()

	// Skip the "import QtQuick.tooling 1.2" preamble.
	for p.cur.kind == tokIdent && p.cur.text == "import" {
		p.skipLine()
		p.next()
	}

	mod := &QMLTypesModule{}
	if p.cur.kind == tokIdent && p.cur.text == "Module" {
		p.next() // eat 'Module'
		p.expect(tokLBrace)
		for p.cur.kind != tokRBrace && p.cur.kind != tokEOF {
			if p.cur.kind == tokIdent && p.cur.text == "Component" {
				comp := p.parseComponent()
				mod.Components = append(mod.Components, comp)
			} else if p.cur.kind == tokIdent {
				p.skipAttribute()
			} else {
				p.next()
			}
		}
		p.expect(tokRBrace)
	}
	return mod, nil
}

func (p *qmltypesParser) parseComponent() QMLTypesComponent {
	p.next() // eat 'Component'
	p.expect(tokLBrace)
	var c QMLTypesComponent
	for p.cur.kind != tokRBrace && p.cur.kind != tokEOF {
		if p.cur.kind != tokIdent {
			p.next()
			continue
		}
		switch p.cur.text {
		case "Property":
			c.Properties = append(c.Properties, p.parseProperty())
		case "Signal":
			c.Signals = append(c.Signals, p.parseSignal())
		case "Method":
			c.Methods = append(c.Methods, p.parseMethod())
		case "Enum":
			c.Enums = append(c.Enums, p.parseEnum())
		case "Component":
			// Nested components exist in some files; parse and append.
			nested := p.parseComponent()
			// Inherit the outer prototype if the nested one is anonymous.
			if nested.Prototype == "" {
				nested.Prototype = c.Name
			}
			// We'll flatten into the module at the caller.
			_ = nested
		default:
			key := p.cur.text
			p.next() // eat key
			if p.cur.kind == tokColon {
				p.next() // eat ':'
				switch key {
				case "name":
					c.Name = p.readString()
				case "file":
					c.File = p.readString()
				case "prototype":
					c.Prototype = p.readString()
				case "defaultProperty":
					c.DefaultProperty = p.readString()
				case "accessSemantics":
					c.AccessSemantics = p.readString()
				case "attachedType":
					c.AttachedType = p.readString()
				case "isSingleton":
					c.IsSingleton = p.readBool()
				case "isCreatable":
					c.IsCreatable = p.readBool()
				case "exports":
					c.Exports = p.readStringArray()
				default:
					p.skipValue()
				}
			}
		}
		p.eatSemicolon()
	}
	p.expect(tokRBrace)
	return c
}

func (p *qmltypesParser) parseProperty() QMLTypesProperty {
	p.next() // eat 'Property'
	p.expect(tokLBrace)
	var prop QMLTypesProperty
	for p.cur.kind != tokRBrace && p.cur.kind != tokEOF {
		if p.cur.kind != tokIdent {
			p.next()
			continue
		}
		key := p.cur.text
		p.next()
		if p.cur.kind == tokColon {
			p.next()
			switch key {
			case "name":
				prop.Name = p.readString()
			case "type":
				prop.Type = p.readString()
			case "isReadonly":
				prop.IsReadonly = p.readBool()
			case "isList":
				prop.IsList = p.readBool()
			case "isPropertyConstant":
				prop.IsPropertyConstant = p.readBool()
			case "notify":
				prop.Notify = p.readString()
			default:
				p.skipValue()
			}
		}
		p.eatSemicolon()
	}
	p.expect(tokRBrace)
	return prop
}

func (p *qmltypesParser) parseSignal() QMLTypesSignal {
	p.next() // eat 'Signal'
	p.expect(tokLBrace)
	var sig QMLTypesSignal
	for p.cur.kind != tokRBrace && p.cur.kind != tokEOF {
		if p.cur.kind == tokIdent && p.cur.text == "Parameter" {
			sig.Parameters = append(sig.Parameters, p.parseParameter())
			p.eatSemicolon()
			continue
		}
		if p.cur.kind != tokIdent {
			p.next()
			continue
		}
		key := p.cur.text
		p.next()
		if p.cur.kind == tokColon {
			p.next()
			if key == "name" {
				sig.Name = p.readString()
			} else {
				p.skipValue()
			}
		}
		p.eatSemicolon()
	}
	p.expect(tokRBrace)
	return sig
}

func (p *qmltypesParser) parseMethod() QMLTypesMethod {
	p.next() // eat 'Method'
	p.expect(tokLBrace)
	var m QMLTypesMethod
	for p.cur.kind != tokRBrace && p.cur.kind != tokEOF {
		if p.cur.kind == tokIdent && p.cur.text == "Parameter" {
			m.Parameters = append(m.Parameters, p.parseParameter())
			p.eatSemicolon()
			continue
		}
		if p.cur.kind != tokIdent {
			p.next()
			continue
		}
		key := p.cur.text
		p.next()
		if p.cur.kind == tokColon {
			p.next()
			switch key {
			case "name":
				m.Name = p.readString()
			case "type":
				m.ReturnType = p.readString()
			default:
				p.skipValue()
			}
		}
		p.eatSemicolon()
	}
	p.expect(tokRBrace)
	return m
}

func (p *qmltypesParser) parseEnum() QMLTypesEnum {
	p.next() // eat 'Enum'
	p.expect(tokLBrace)
	var e QMLTypesEnum
	for p.cur.kind != tokRBrace && p.cur.kind != tokEOF {
		if p.cur.kind != tokIdent {
			p.next()
			continue
		}
		key := p.cur.text
		p.next()
		if p.cur.kind == tokColon {
			p.next()
			switch key {
			case "name":
				e.Name = p.readString()
			case "isFlag":
				e.IsFlag = p.readBool()
			case "values":
				e.Values = p.readEnumValues()
			default:
				p.skipValue()
			}
		}
		p.eatSemicolon()
	}
	p.expect(tokRBrace)
	return e
}

func (p *qmltypesParser) parseParameter() QMLTypesParameter {
	p.next() // eat 'Parameter'
	p.expect(tokLBrace)
	var param QMLTypesParameter
	for p.cur.kind != tokRBrace && p.cur.kind != tokEOF {
		if p.cur.kind != tokIdent {
			p.next()
			continue
		}
		key := p.cur.text
		p.next()
		if p.cur.kind == tokColon {
			p.next()
			switch key {
			case "name":
				param.Name = p.readString()
			case "type":
				param.Type = p.readString()
			default:
				p.skipValue()
			}
		}
		p.eatSemicolon()
	}
	p.expect(tokRBrace)
	return param
}

// ----- Value readers -----

func (p *qmltypesParser) readString() string {
	if p.cur.kind == tokString {
		s := p.cur.text
		p.next()
		return s
	}
	// Unquoted identifier used as a string value.
	if p.cur.kind == tokIdent {
		s := p.cur.text
		p.next()
		return s
	}
	return ""
}

func (p *qmltypesParser) readBool() bool {
	if p.cur.kind == tokTrue {
		p.next()
		return true
	}
	if p.cur.kind == tokFalse {
		p.next()
		return false
	}
	s := p.readString()
	return s == "true"
}

func (p *qmltypesParser) readStringArray() []string {
	if p.cur.kind != tokLBracket {
		// Single value, not an array.
		return []string{p.readString()}
	}
	p.next() // eat '['
	var arr []string
	for p.cur.kind != tokRBracket && p.cur.kind != tokEOF {
		if p.cur.kind == tokComma {
			p.next()
			continue
		}
		arr = append(arr, p.readString())
	}
	p.expect(tokRBracket)
	return arr
}

// readEnumValues handles both formats:
//   - Qt6 string array: ["Value1", "Value2"]
//   - Qt5 object map:   { "Value1": 0, "Value2": 1 }
func (p *qmltypesParser) readEnumValues() []string {
	if p.cur.kind == tokLBracket {
		return p.readStringArray()
	}
	if p.cur.kind == tokLBrace {
		p.next()
		var vals []string
		for p.cur.kind != tokRBrace && p.cur.kind != tokEOF {
			if p.cur.kind == tokComma {
				p.next()
				continue
			}
			if p.cur.kind == tokString {
				vals = append(vals, p.cur.text)
				p.next()
				if p.cur.kind == tokColon {
					p.next()
					p.skipValue() // skip the numeric value
				}
			} else {
				p.next()
			}
		}
		p.expect(tokRBrace)
		return vals
	}
	return nil
}

func (p *qmltypesParser) skipAttribute() {
	p.next() // eat key
	if p.cur.kind == tokColon {
		p.next()
		p.skipValue()
	}
	p.eatSemicolon()
}

func (p *qmltypesParser) skipValue() {
	switch p.cur.kind {
	case tokLBrace:
		p.skipBlock(tokLBrace, tokRBrace)
	case tokLBracket:
		p.skipBlock(tokLBracket, tokRBracket)
	default:
		p.next()
	}
}

func (p *qmltypesParser) skipBlock(open, close tokKind) {
	depth := 1
	p.next() // eat opening
	for depth > 0 && p.cur.kind != tokEOF {
		switch p.cur.kind {
		case open:
			depth++
		case close:
			depth--
		}
		p.next()
	}
}

func (p *qmltypesParser) skipLine() {
	for p.pos < len(p.src) && p.src[p.pos] != '\n' {
		p.pos++
	}
}

func (p *qmltypesParser) eatSemicolon() {
	if p.cur.kind == tokSemicolon {
		p.next()
	}
}

func (p *qmltypesParser) expect(kind tokKind) {
	if p.cur.kind == kind {
		p.next()
	}
}

// ----- Lexer -----

func (p *qmltypesParser) next() {
	p.skipWhitespaceAndComments()
	if p.pos >= len(p.src) {
		p.cur = token{kind: tokEOF}
		return
	}

	ch := p.src[p.pos]
	switch ch {
	case '{':
		p.cur = token{kind: tokLBrace, text: "{"}
		p.pos++
	case '}':
		p.cur = token{kind: tokRBrace, text: "}"}
		p.pos++
	case '[':
		p.cur = token{kind: tokLBracket, text: "["}
		p.pos++
	case ']':
		p.cur = token{kind: tokRBracket, text: "]"}
		p.pos++
	case ':':
		p.cur = token{kind: tokColon, text: ":"}
		p.pos++
	case ',':
		p.cur = token{kind: tokComma, text: ","}
		p.pos++
	case ';':
		p.cur = token{kind: tokSemicolon, text: ";"}
		p.pos++
	case '"':
		p.cur = token{kind: tokString, text: p.readQuotedString()}
	default:
		if ch == '-' || ch == '+' || (ch >= '0' && ch <= '9') {
			p.cur = token{kind: tokNumber, text: p.readNumber()}
		} else if isIdentStart(ch) {
			ident := p.readIdent()
			switch ident {
			case "true":
				p.cur = token{kind: tokTrue, text: ident}
			case "false":
				p.cur = token{kind: tokFalse, text: ident}
			default:
				p.cur = token{kind: tokIdent, text: ident}
			}
		} else {
			p.pos++
			p.next()
		}
	}
}

func (p *qmltypesParser) skipWhitespaceAndComments() {
	for p.pos < len(p.src) {
		ch := p.src[p.pos]
		if ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' {
			p.pos++
			continue
		}
		if ch == '/' && p.pos+1 < len(p.src) {
			if p.src[p.pos+1] == '/' {
				for p.pos < len(p.src) && p.src[p.pos] != '\n' {
					p.pos++
				}
				continue
			}
			if p.src[p.pos+1] == '*' {
				p.pos += 2
				for p.pos+1 < len(p.src) {
					if p.src[p.pos] == '*' && p.src[p.pos+1] == '/' {
						p.pos += 2
						break
					}
					p.pos++
				}
				continue
			}
		}
		break
	}
}

func (p *qmltypesParser) readQuotedString() string {
	p.pos++ // skip opening "
	var b strings.Builder
	for p.pos < len(p.src) {
		ch := p.src[p.pos]
		if ch == '\\' && p.pos+1 < len(p.src) {
			p.pos++
			b.WriteByte(p.src[p.pos])
			p.pos++
			continue
		}
		if ch == '"' {
			p.pos++
			return b.String()
		}
		b.WriteByte(ch)
		p.pos++
	}
	return b.String()
}

func (p *qmltypesParser) readNumber() string {
	start := p.pos
	if p.src[p.pos] == '-' || p.src[p.pos] == '+' {
		p.pos++
	}
	for p.pos < len(p.src) && (p.src[p.pos] >= '0' && p.src[p.pos] <= '9' || p.src[p.pos] == '.') {
		p.pos++
	}
	// Hex: 0x...
	if p.pos-start >= 2 && p.src[start] == '0' && (p.src[start+1] == 'x' || p.src[start+1] == 'X') {
		for p.pos < len(p.src) && isHexDigit(p.src[p.pos]) {
			p.pos++
		}
	}
	return p.src[start:p.pos]
}

func (p *qmltypesParser) readIdent() string {
	start := p.pos
	for p.pos < len(p.src) && isQmltypesIdentChar(p.src[p.pos]) {
		p.pos++
	}
	// Allow C++ scoped names like "std::vector<bool>" and "QQuickItem::TransformOrigin".
	// Only consume `::` (double colon) — a single `:` is the attribute separator.
	for p.pos+1 < len(p.src) && ((p.src[p.pos] == ':' && p.src[p.pos+1] == ':') || p.src[p.pos] == '<' || p.src[p.pos] == '>') {
		if p.src[p.pos] == ':' {
			p.pos += 2 // skip ::
		} else {
			p.pos++ // skip < or >
		}
		for p.pos < len(p.src) && isQmltypesIdentChar(p.src[p.pos]) {
			p.pos++
		}
	}
	return p.src[start:p.pos]
}

func isIdentStart(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isQmltypesIdentChar(ch byte) bool {
	return isIdentStart(ch) || (ch >= '0' && ch <= '9') || ch == '.'
}

func isHexDigit(ch byte) bool {
	return (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

// cppTypeToQML maps common C++ type names to QML-friendly names for display.
func cppTypeToQML(t string) string {
	switch t {
	case "double", "float", "qreal":
		return "real"
	case "int", "qint32":
		return "int"
	case "bool":
		return "bool"
	case "QString":
		return "string"
	case "QColor":
		return "color"
	case "QUrl":
		return "url"
	case "QVariant", "QJSValue":
		return "var"
	case "QQmlComponent":
		return "Component"
	case "QQmlListProperty":
		return "list"
	case "QPointF", "QPoint":
		return "point"
	case "QSizeF", "QSize":
		return "size"
	case "QRectF", "QRect":
		return "rect"
	case "QVector2D":
		return "vector2d"
	case "QVector3D":
		return "vector3d"
	case "QVector4D":
		return "vector4d"
	case "QQuaternion":
		return "quaternion"
	case "QMatrix4x4":
		return "matrix4x4"
	case "QFont":
		return "font"
	case "QDate":
		return "date"
	case "QDateTime":
		return "date"
	case "":
		return "void"
	}
	// Strip QQuick prefix and pointer suffix for readability.
	s := strings.TrimSuffix(t, "*")
	s = strings.TrimPrefix(s, "QQuick")
	s = strings.TrimPrefix(s, "Q")
	if s == "" {
		return t
	}
	return s
}

// qtVersionFromExport extracts a numeric sort key from an export version
// string like "6.7" → 607.
func qtVersionFromExport(exp string) int {
	if i := lastIndexByte(exp, ' '); i >= 0 {
		ver := exp[i+1:]
		parts := strings.SplitN(ver, ".", 2)
		major, _ := strconv.Atoi(parts[0])
		minor := 0
		if len(parts) > 1 {
			minor, _ = strconv.Atoi(parts[1])
		}
		return major*100 + minor
	}
	return 0
}

