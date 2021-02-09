package canvas

import (
	"unicode/utf8"
)

// EmojiLexer splits the text and the emojis from s.
type EmojiLexer func(s string, text func(string), emoji func(string))

// NewEmojiTextBox is an advanced text formatter that will calculate text placement based on the settings. It takes a font face, a string, the width or height of the box (can be zero for no limit), horizontal and vertical alignment (Left, Center, Right, Top, Bottom or Justify), text indentation for the first line and line stretch (percentage to stretch the line based on the line height).
func NewEmojiTextBox(ff FontFace, s string, width, height float64, halign, valign TextAlign, indent, lineStretch float64, lexer EmojiLexer) (*Text, []Emoji) {
	return NewRichEmojiText(lexer).Add(ff, s).ToText(width, height, halign, valign, indent, lineStretch)
}

// RichEmojiText allows to build up a rich text with text spans of different font faces and by fitting that into a box.
type RichEmojiText struct {
	*RichText
	lexer EmojiLexer
}

// NewRichEmojiText returns a new RichEmojiText.
func NewRichEmojiText(lexer EmojiLexer) *RichEmojiText {
	return &RichEmojiText{
		RichText: NewRichText(),
		lexer:    lexer,
	}
}

// Add adds a new text span element.
func (rt *RichEmojiText) Add(ff FontFace, s string) *RichEmojiText {
	if 0 < len(s) {
		rPrev := ' '
		rNext, size := utf8.DecodeRuneInString(s)
		if 0 < len(rt.text) {
			rPrev, _ = utf8.DecodeLastRuneInString(rt.text)
		}
		if isWhitespace(rPrev) && isWhitespace(rNext) {
			s = s[size:]
		}
	}

	start := len(rt.text)
	rt.text += s

	rt.lexer(s, func(txt string) {
		// TODO: can we simplify this? Just merge adjacent spans, don't split at newlines or sentences?
		i := 0
		for _, boundary := range calcTextBoundaries(txt, 0, len(txt)) {
			if boundary.kind == lineBoundary || boundary.kind == sentenceBoundary || boundary.kind == eofBoundary {
				j := boundary.pos + boundary.size
				if i < j {
					extendPrev := false
					if i == 0 && boundary.kind != lineBoundary && 0 < len(rt.spans) && rt.spans[len(rt.spans)-1].Face.Equals(ff) {
						prevSpan := rt.spans[len(rt.spans)-1]
						if 1 < len(prevSpan.boundaries) {
							prevBoundaryKind := prevSpan.boundaries[len(prevSpan.boundaries)-2].kind
							if prevBoundaryKind != lineBoundary && prevBoundaryKind != sentenceBoundary {
								extendPrev = true
							}
						} else if !prevSpan.IsEmoji {
							extendPrev = true
						}
					}

					if extendPrev {
						diff := len(rt.spans[len(rt.spans)-1].Text)
						rt.spans[len(rt.spans)-1] = newTextSpan(ff, rt.text[:start+j], start+i-diff)
					} else {
						rt.spans = append(rt.spans, newTextSpan(ff, rt.text[:start+j], start+i))
					}
				}
				i = j
			}
		}
		start += len(txt)
	}, func(emj string) {
		rt.spans = append(rt.spans, newEmojiSpan(ff, emj))
		start += len(emj)
		// fmt.Println(emj)
	})
	// if strings.HasSuffix(s, "🏃‍♀️") {
	// 	fmt.Printf("%#v", rt.spans)
	// }
	rt.fonts[ff.Font] = true
	return rt
}

func newEmojiSpan(ff FontFace, emj string) TextSpan {
	return TextSpan{
		Face:            ff,
		Text:            emj,
		IsEmoji:         true,
		width:           ff.Size * ff.Scale * .9,
		boundaries:      []textBoundary{{eofBoundary, len(emj), 0}},
		dx:              0.0,
		SentenceSpacing: 0.0,
		WordSpacing:     0.0,
		GlyphSpacing:    0.0,
	}
}

// ToText takes the added text spans and fits them within a given box of certain width and height.
func (rt *RichEmojiText) ToText(width, height float64, halign, valign TextAlign, indent, lineStretch float64) (*Text, []Emoji) {
	text := rt.RichText.ToText(width, height, halign, valign, indent, lineStretch)
	if text == nil || len(text.lines) == 0 {
		return text, nil
	}

	var emjs []Emoji
	for i, l := range text.lines {
		for j, s := range l.spans {
			if s.IsEmoji {
				emjs = append(emjs, Emoji{
					Text:  s.Text,
					X:     s.dx + s.Face.Size*s.Face.Scale*.05,
					Y:     l.y,
					Scale: s.Face.Size * s.Face.Scale,
				})

				s.Text = ""
				text.lines[i].spans[j] = s
			}
		}
	}

	return text, emjs
}

type Emoji struct {
	Text  string
	X, Y  float64
	Scale float64
}
