package shell

import (
	"errors"
	"regexp"
	"strings"

	"paleotronic.com/fmt"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
	//	"paleotronic.com/utils"
)

type PlusRegMatcher struct {
	RegExp     *regexp.Regexp // eg: ^(.+)?[/]?([^/]+)?$
	RegToToken []*types.Token // $1:, $2:*.*
}

type PlusWrapper struct {
	dialect.Command
	WrappedName    string
	WrappedCommand interfaces.Functioner
	WrappedParam   []*types.Token
	Matchers       []PlusRegMatcher
}

var plusWrapToken = regexp.MustCompile("^[$]([0-9]+)$")
var plusEvalToken = regexp.MustCompile("^[$][{](.+)[}]$")

func eval(ent interfaces.Interpretable, pattern string) string {

	if plusEvalToken.MatchString(pattern) {
		m := plusEvalToken.FindAllStringSubmatch(pattern, -1)
		switch m[0][1] {
		case "cwd":
			return ent.GetWorkDir()
		case "slot":
			return utils.IntToStr(ent.GetMemIndex())
		default:
			return "undef"
		}
	}
	return pattern

}

func (this *PlusWrapper) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* enforce non void return */
	fmt.Printf("PREPARSE: About to execute command %s: %s\n", this.WrappedName, strings.Join(tokens.Strings(), ", "))

	if this.Matchers != nil && len(this.Matchers) > 0 {
		joined := strings.Join(tokens.Strings(), "")

		for i, matcher := range this.Matchers {

			if matcher.RegExp.MatchString(joined) {
				// We matched the pattern
				newtokens := *types.NewTokenList()
				m := matcher.RegExp.FindAllStringSubmatch(joined, -1)
				for _, tpat := range matcher.RegToToken {

					if newtokens.Size() > 0 {
						newtokens.Push(types.NewToken(types.SEPARATOR, ","))
					}

					tmp := strings.Split(tpat.Content, ":")
					for len(tmp) < 2 {
						tmp = append(tmp, "")
					}
					var t *types.Token = types.NewToken(tpat.Type, eval(caller, tmp[1]))
					if plusWrapToken.MatchString(tmp[0]) {
						mm := plusWrapToken.FindAllStringSubmatch(tmp[0], -1)
						index := utils.StrToInt(mm[0][1])
						t.Content = m[0][index]
					}
					newtokens.Push(t)
				}

				fmt.Printf("Set pattern %d tokens to [%s]\n", i, strings.Join(newtokens.Strings(), ", "))
				tokens = newtokens
				break

			}
		}
	}

	if len(this.WrappedParam) != 0 {
		nt := *types.NewTokenList()

		for _, tt := range this.WrappedParam {
			if tt.Type == types.PLACEHOLDER {

				i := tt.AsInteger()

				if i == 0 {
					nt.Content = append(nt.Content, tokens.Content...)
				} else if i >= 1 && i <= tokens.Size() {
					nt.Content = append(nt.Content, tokens.Get(i-1))
				}

			} else {
				nt.Add(tt)
			}
		}

		tokens.Content = nt.Content
	}

	if this.WrappedCommand == nil {
		return 0, errors.New("NULL command error")
	}

	//return this.WrappedCommand.Execute(env, caller, tokens, Scope, LPC)

	this.WrappedCommand.SetEntity(caller)
	this.WrappedCommand.GetStack().Clear()
	fmt.Printf("Named mode for %s is %v\n", this.WrappedName, this.WrappedCommand.GetRaw())
	fmt.Printf("About to execute command %s: %s\n", this.WrappedName, strings.Join(tokens.Strings(), ", "))
	this.WrappedCommand.FunctionExecute(&tokens)

	return 0, nil

}

func (this *PlusWrapper) Syntax() string {

	/* vars */
	var result string

	result = this.WrappedName

	/* enforce non void return */
	return result

}
