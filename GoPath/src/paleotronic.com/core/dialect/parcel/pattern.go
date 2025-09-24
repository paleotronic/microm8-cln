package parcel

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"strings"

	"paleotronic.com/log"
)

/*
pattern DIGIT			[0-9]
pattern ALPHA  			[A-Za-z]
pattern IDENTIFIER		{ALPHA}({DIGIT}|{ALPHA})+
pattern WHITESPACE  	[ \t\r\n]
pattern INTEGER     	[-]?{DIGIT}+
pattern DEREFERENCE 	[:]{IDENTIFIER}
pattern STRING      	[^ \t\r\n]
pattern STRINGLITERAL 	["]{STRING}
*/

var pr = regexp.MustCompile("^[ \t]*pattern[ \t]+([A-Za-z0-9]+)[ \t]+(.+)$")
var tk = regexp.MustCompile("^[ \t]*token[ \t]+([.]?[A-Za-z0-9]+)[ \t]+(.+)$")
var sp = regexp.MustCompile("[{][A-Za-z_]+[}]")
var ru = regexp.MustCompile("^[ \t]*rule[ \t]+([A-Za-z0-9]+)[ \t]+(.+)$")

type Pattern struct {
	r       *regexp.Regexp
	Label   string
	Content string
}

type Token struct {
	Label string
	p     *Pattern
}

type Rule struct {
	Label string
	r     *regexp.Regexp
}

type Lexer struct {
	Patterns   []*Pattern
	patternIdx map[string]int
	Tokens     []*Token
	tokenIdx   map[string]int
	Rules      []*Rule
	ruleIdx    map[string]int
	driver     GrammarDriver
}

type PatternMatchResult struct {
	Start   int
	End     int
	Length  int
	Pattern *Pattern
}

type TokenMatchResult struct {
	Start   int
	End     int
	Length  int
	Token   *Token
	Ignore  bool
	Content string
}

type RuleMatchResult struct {
	Start       int
	End         int
	Length      int
	Rule        *Rule
	Tokens      []*TokenMatchResult
	PatternForm string
}

func NewLexer(driver GrammarDriver) *Lexer {
	return &Lexer{
		Patterns:   []*Pattern{},
		patternIdx: map[string]int{},
		Tokens:     []*Token{},
		tokenIdx:   map[string]int{},
		Rules:      []*Rule{},
		ruleIdx:    map[string]int{},
		driver:     driver,
	}
}

func (pm *Lexer) LoadString(s string) error {
	return pm.Load(bytes.NewBuffer([]byte(s)))
}

func (pm *Lexer) Load(r io.Reader) error {

	raw, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	lines := strings.Split(string(raw), "\n")
	for lno, line := range lines {
		line = strings.Trim(line, "\t\r ")
		if line == "" {
			continue
		}
		if pr.MatchString(line) {
			m := pr.FindAllStringSubmatch(line, -1)
			pattern := m[0][1]
			def := m[0][2]
			log.Printf("Pattern def %s -> %s", pattern, def)
			err = pm.AddPattern(pattern, def)
			if err != nil {
				return fmt.Errorf("Problem processing pattern at line %d: %v", lno, err)
			}
		} else if tk.MatchString(line) {
			m := tk.FindAllStringSubmatch(line, -1)
			pattern := m[0][1]
			def := m[0][2]
			log.Printf("Token def %s -> %s", pattern, def)
			err = pm.AddToken(pattern, def)
			if err != nil {
				return fmt.Errorf("Problem processing token def at line %d: %v", lno, err)
			}
		} else if ru.MatchString(line) {
			m := ru.FindAllStringSubmatch(line, -1)
			pattern := m[0][1]
			def := m[0][2]
			log.Printf("Rule def %s -> %s", pattern, def)
			err = pm.AddRule(pattern, def)
			if err != nil {
				return fmt.Errorf("Problem processing rule def at line %d: %v", lno, err)
			}
		}
	}

	return nil

}

func (pm *Lexer) GetPattern(patternName string) (int, *Pattern) {
	idx, ok := pm.patternIdx[patternName]
	if !ok {
		return -1, nil
	}
	return idx, pm.Patterns[idx]
}

func (pm *Lexer) GetToken(tokenName string) (int, *Token) {
	idx, ok := pm.tokenIdx[tokenName]
	if !ok {
		return -1, nil
	}
	return idx, pm.Tokens[idx]
}

func (pm *Lexer) GetRule(ruleName string) (int, *Rule) {
	idx, ok := pm.ruleIdx[ruleName]
	if !ok {
		return -1, nil
	}
	return idx, pm.Rules[idx]
}

func (pm *Lexer) BuildPattern(patternName string, patternDef string) (*Pattern, error) {
	var loc []int
	for sp.MatchString(patternDef) {
		loc = sp.FindStringIndex(patternDef)
		subPattern := strings.Trim(patternDef[loc[0]:loc[1]], "{}")
		log.Printf("- pattern contains sub pattern %s", subPattern)
		_, sp := pm.GetPattern(subPattern)
		if sp == nil {
			return nil, fmt.Errorf("sub pattern is not defined: {%s}", subPattern)
		}
		patternDef = strings.Replace(patternDef, "{"+subPattern+"}", sp.Content, -1)
	}

	log.Printf(" - pattern %s resolves to regexp %s", patternName, patternDef)

	r, err := regexp.Compile(patternDef)
	if err != nil {
		return nil, err
	}

	pd := &Pattern{
		r:       r,
		Label:   patternName,
		Content: patternDef,
	}
	return pd, nil
}

func (pm *Lexer) AddPattern(patternName string, patternDef string) error {
	nextIdx := len(pm.Patterns)

	// existing pattern
	exIdx, _ := pm.GetPattern(patternName)
	if exIdx != -1 {
		nextIdx = exIdx
	}

	pd, err := pm.BuildPattern(patternName, patternDef)
	if err != nil {
		return err
	}

	pm.Patterns = append(pm.Patterns, pd)
	pm.patternIdx[patternName] = nextIdx

	return nil
}

func (pm *Lexer) AddToken(patternName string, patternDef string) error {
	nextIdx := len(pm.Tokens)

	// existing token
	exIdx, _ := pm.GetToken(patternName)
	if exIdx != -1 {
		nextIdx = exIdx
	}

	pd, err := pm.BuildPattern(patternName, "(?i)"+patternDef)
	if err != nil {
		return err
	}

	pm.Tokens = append(pm.Tokens,
		&Token{
			Label: patternName,
			p:     pd,
		},
	)
	pm.tokenIdx[patternName] = nextIdx

	return nil
}

func (pm *Lexer) AddRule(patternName string, patternDef string) error {
	nextIdx := len(pm.Rules)

	r, err := regexp.Compile(patternDef)
	if err != nil {
		return err
	}

	pm.Rules = append(pm.Rules,
		&Rule{
			Label: patternName,
			r:     r,
		},
	)
	pm.ruleIdx[patternName] = nextIdx

	return nil
}

func (pm *Lexer) PatternMatches(chunk string, mustAnchor bool, allMatches bool) []PatternMatchResult {
	var results = []PatternMatchResult{}
	for _, p := range pm.Patterns {
		if p.r.MatchString(chunk) {
			loc := p.r.FindStringIndex(chunk)
			if mustAnchor && loc[0] != 0 {
				continue
			}
			results = append(results,
				PatternMatchResult{
					Start:   loc[0],
					End:     loc[1],
					Length:  loc[1] - loc[0],
					Pattern: p,
				},
			)
			if !allMatches {
				return results
			}
		}
	}
	return results
}

func (pm *Lexer) GetLongestPatternMatch(chunk string, mustAnchor bool) *PatternMatchResult {
	matches := pm.PatternMatches(chunk, mustAnchor, true)
	if len(matches) == 0 {
		return nil
	}
	if len(matches) == 1 {
		return &matches[0]
	}
	var idx = -1
	var maxlen = 0
	for i, m := range matches {
		if m.Length > maxlen {
			idx = i
		}
	}
	return &matches[idx]
}

func (pm *Lexer) TokenMatches(chunk string, startPos int, mustAnchor bool, allMatches bool) []TokenMatchResult {
	var results = []TokenMatchResult{}
	if startPos >= len(chunk) {
		return results
	}
	chunk = chunk[startPos:]

	for _, t := range pm.Tokens {
		//log.Printf("Matching against token type: [%s]", t.Label)
		if t.p.r.MatchString(chunk) {
			loc := t.p.r.FindStringIndex(chunk)
			if mustAnchor && loc[0] != 0 {
				//log.Println("Skipping non-anchored")
				continue
			}
			tmr := TokenMatchResult{
				Start:   loc[0] + startPos,
				End:     loc[1] + startPos,
				Length:  loc[1] - loc[0],
				Token:   t,
				Ignore:  strings.HasPrefix(t.Label, "."),
				Content: chunk[loc[0]:loc[1]],
			}
			results = append(results, tmr)
			if !allMatches {
				return results
			}
		}
	}
	return results
}

func (pm *Lexer) GetLongestTokenMatch(chunk string, startPos int, mustAnchor bool) *TokenMatchResult {
	matches := pm.TokenMatches(chunk, startPos, mustAnchor, true)
	if len(matches) == 0 {
		return nil
	}
	if len(matches) == 1 {
		return &matches[0]
	}
	var idx = -1
	var maxlen = 0
	for i, m := range matches {
		if m.Length > maxlen {
			idx = i
			maxlen = m.Length
		}
	}
	return &matches[idx]
}

func (pm *Lexer) Tokenize(chunk string) ([]*TokenMatchResult, error) {
	var r = []*TokenMatchResult{}
	pos := 0

	m := pm.GetLongestTokenMatch(chunk, pos, true)
	for m != nil && pos < len(chunk) {
		if !m.Ignore {
			r = append(r, m)
		}
		pos = m.End
		m = pm.GetLongestTokenMatch(chunk, pos, true)
	}

	if pos < len(chunk) {
		return r, fmt.Errorf("unrecognized token at pos %d", pos)
	}

	return r, nil
}

func (pm *Lexer) TokenizeToString(chunk string) (string, error) {
	r, err := pm.Tokenize(chunk)
	var tokens = make([]string, len(r))
	for i, t := range r {
		tokens[i] = t.Token.Label
	}
	return strings.Join(tokens, " "), err
}

func (pm *Lexer) RuleMatches(chunk string, startPos int, allMatches bool) []RuleMatchResult {
	var results = []RuleMatchResult{}
	if startPos >= len(chunk) {
		return results
	}
	chunk = chunk[startPos:]

	r, err := pm.Tokenize(chunk)
	if err != nil {
		return results
	}
	var tokens = make([]string, len(r))
	for i, t := range r {
		tokens[i] = t.Token.Label
	}
	pattern := strings.Join(tokens, " ")

	for _, t := range pm.Rules {
		//log.Printf("Matching against token type: [%s]", t.Label)
		if t.r.MatchString(pattern) {
			loc := t.r.FindStringIndex(pattern)
			tmp := pattern[loc[0]:loc[1]]
			parts := strings.Split(tmp, " ")
			results = append(results,
				RuleMatchResult{
					Start:       loc[0] + startPos,
					End:         loc[1] + startPos,
					Length:      loc[1] - loc[0],
					Rule:        t,
					Tokens:      r[0:len(parts)],
					PatternForm: tmp,
				},
			)
			if !allMatches {
				return results
			}
		}
	}
	return results
}

func (pm *Lexer) GetLongestRuleMatch(chunk string, startPos int) *RuleMatchResult {
	matches := pm.RuleMatches(chunk, startPos, false)
	if len(matches) == 0 {
		return nil
	}
	if len(matches) == 1 {
		return &matches[0]
	}
	var idx = -1
	var maxlen = 0
	for i, m := range matches {
		if m.Length > maxlen {
			idx = i
			maxlen = m.Length
		}
	}
	return &matches[idx]
}

func (pm *Lexer) RuleMatchesFromTokens(r []*TokenMatchResult, startPos int, allMatches bool) []RuleMatchResult {
	var results = []RuleMatchResult{}
	if startPos >= len(r) {
		return results
	}
	r = r[startPos:]

	var tokens = make([]string, len(r))
	for i, t := range r {
		tokens[i] = t.Token.Label + "(" + strings.ToLower(t.Content) + ")"
	}
	pattern := strings.Join(tokens, " ")
	log.Printf("trying rule match for -> %s", pattern)

	for _, t := range pm.Rules {
		//log.Printf("Matching against token type: [%s]", t.Label)
		if t.r.MatchString(pattern) {
			loc := t.r.FindStringIndex(pattern)
			tmp := pattern[loc[0]:loc[1]]
			parts := strings.Split(tmp, " ")
			results = append(results,
				RuleMatchResult{
					Start:       startPos,
					End:         startPos + len(parts),
					Length:      len(parts),
					Rule:        t,
					Tokens:      r[startPos : startPos+len(parts)],
					PatternForm: tmp,
				},
			)
			if !allMatches {
				return results
			}
		}
	}
	return results
}

func (pm *Lexer) ProcessInputStream(s string) error {

	var err error

	if pm.driver != nil {
		err = pm.driver.OnBeginStream(pm)
	}

	pos := 0
	for err == nil && pos < len(s) {
		tmr := pm.GetLongestTokenMatch(s, pos, true)
		if tmr != nil {
			if pm.driver != nil {
				pm.driver.OnTokenRecognized(tmr)
			}
			pos = tmr.End // move to end of this token
		} else {
			x := len(s)
			if x-pos > 20 {
				x = pos + 20
			}
			err = fmt.Errorf("unrecognized token: %s", s[pos:x])
			if pm.driver != nil {
				err = pm.driver.OnTokenUnrecognized(s, pos, err)
			}
		}
	}

	if pm.driver != nil && err == nil {
		err = pm.driver.OnEndStream(pm)
	}

	return err

}
