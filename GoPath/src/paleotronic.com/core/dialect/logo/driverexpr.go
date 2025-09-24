package logo

import (
	"errors"
	"strings"

	"paleotronic.com/log"

	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

func (d *LogoDriver) Compare(a *types.Token, b *types.Token) (bool, int) {

	if a == nil && b == nil {
		return true, 0
	}
	if a == nil {
		return false, -1
	}
	if b == nil {
		return false, 1
	}

	log.Printf("[compare-entry] a.content = %s(%s), b.content = %s(%s)", a.Type.String(), a.Content, b.Type.String(), b.Content)

	if a.Type != b.Type {
		if !a.IsIn([]types.TokenType{types.WORD, types.IDENTIFIER, types.STRING}) && !b.IsIn([]types.TokenType{types.WORD, types.IDENTIFIER, types.STRING}) {
			if !a.IsNumeric() || !b.IsNumeric() {
				return false, 0
			}
		}
	}
	if a.Type == types.LIST && b.Type == types.LIST {
		// do list comparision
		if a.List.Size() < b.List.Size() {
			return false, -1
		} else if b.List.Size() < a.List.Size() {
			return false, 1
		}
		// size is the same, compare elements until different
		for i := 0; i < a.List.Size(); i++ {
			if same, diff := d.Compare(a.List.Content[i], b.List.Content[i]); !same {
				return same, diff
			}
		}

		return true, 0
	}

	// not a list and the same type
	switch a.Type {
	case types.STRING, types.WORD, types.IDENTIFIER:
		// string compare
		log.Printf("[compare] a.content = %s, b.content = %s", a.Content, b.Content)
		if a.Content == b.Content {
			return true, 0
		}
		if a.Content < b.Content {
			return false, -1
		}
		return false, 1
	default:
		log.Printf("[compare] a.content = %s(%s), b.content = %s(%s)", a.Type.String(), a.Content, b.Type.String(), b.Content)
		// numeric compare
		aa := a.AsExtended()
		bb := b.AsExtended()
		if aa == bb {
			return true, 0
		}
		if aa < bb {
			return false, -1
		}
		return false, 1
	}
}

type opSearch struct {
	Kind  types.TokenType
	Match string
	Skip  bool
}

func (d *LogoDriver) CollapseSimpleExpr(tokens *types.TokenList) (*types.TokenList, error) {

	log.Printf("[collapse] %s", tlistStr("", tokens))

	tmp := tokens.Copy()

	var pos int
	var searches = []*opSearch{
		&opSearch{types.OPERATOR, "/", false},
		&opSearch{types.OPERATOR, "*", false},
		&opSearch{types.OPERATOR, "", false},
		&opSearch{types.COMPARITOR, "", false},
		&opSearch{types.ASSIGNMENT, "", false},
	}

	var nextOp = func(start int, s *opSearch) {
		pos = tmp.IndexOfN(start, s.Kind, s.Match)
	}

	for _, s := range searches {

		nextOp(0, s)

		for pos > 0 && pos < tmp.Size()-1 {
			log.Printf("[collapse] op found at pos %d [%+v]", pos, tmp.AsString())
			op := tmp.Get(pos)
			l := tmp.Get(pos - 1)
			r := tmp.Get(pos + 1)
			if l.Type != types.VARIABLE && l.Type != types.NUMBER && l.Type != types.INTEGER && l.Type != types.EXPRESSIONLIST && l.Type != types.LIST {
				log.Println("skipping left not number or var + ", l.Type.String())
				nextOp(pos+1, s)
				continue
			}
			if r.Type != types.VARIABLE && r.Type != types.NUMBER && r.Type != types.INTEGER && r.Type != types.EXPRESSIONLIST && r.Type != types.LIST {
				log.Println("skipping right not number or var")
				nextOp(pos+1, s)
				continue
			}
			if l.Type == types.VARIABLE {
				log.Printf("[collapse] left hand expr is var")
				name := l.Content
				l, _ = d.GetVar(l.Content)
				if l == nil {
					return nil, errors.New("No such variable " + name)
				}
			} else if l.Type == types.EXPRESSIONLIST {
				tmp, err := d.ParseExprRLCollapse(l.List, true)
				if err != nil {
					return nil, err
				}
				l = tmp[0]
			}
			if r.Type == types.VARIABLE {
				log.Printf("[collapse] right hand expr is var")
				name := r.Content
				r, _ = d.GetVar(r.Content)
				if r == nil {
					return nil, errors.New("No such variable " + name)
				}
			} else if r.Type == types.EXPRESSIONLIST {
				tmp, err := d.ParseExprRLCollapse(r.List, true)
				if err != nil {
					return nil, err
				}
				r = tmp[0]
			}

			var v *types.Token
			var isSame bool
			var diffDir int

			if op.Type == types.COMPARITOR || op.Type == types.ASSIGNMENT {
				isSame, diffDir = d.Compare(l, r)
			}

			switch op.Content {
			case "<=":
				cv := "0"
				if isSame || diffDir == -1 {
					cv = "1"
				}
				v = types.NewToken(types.NUMBER, cv)
			case ">=":
				cv := "0"
				if isSame || diffDir == 1 {
					cv = "1"
				}
				v = types.NewToken(types.NUMBER, cv)
			case ">":
				cv := "0"
				if diffDir == 1 {
					cv = "1"
				}
				v = types.NewToken(types.NUMBER, cv)
			case "<":
				cv := "0"
				if diffDir == -1 {
					cv = "1"
				}
				v = types.NewToken(types.NUMBER, cv)
			case "=":
				cv := "0"
				if isSame {
					cv = "1"
				}
				v = types.NewToken(types.NUMBER, cv)
			case "+":
				v = types.NewToken(types.NUMBER, utils.FloatToStr(l.AsExtended()+r.AsExtended()))
			case "-":
				v = types.NewToken(types.NUMBER, utils.FloatToStr(l.AsExtended()-r.AsExtended()))
			case "*":
				v = types.NewToken(types.NUMBER, utils.FloatToStr(l.AsExtended()*r.AsExtended()))
			case "/":
				if r.AsExtended() != 0 {
					v = types.NewToken(types.NUMBER, utils.FloatToStr(l.AsExtended()/r.AsExtended()))
				} else {
					return tokens, errors.New("division by zero")
				}
			default:
				log.Println("Unknown operator " + op.Content)
				return tokens, errors.New("unknown operator")
			}
			log.Printf("[collapse] %f %s %f ==> %f", l.AsExtended(), op.Content, r.AsExtended(), v.AsExtended())
			b, e := tmp.Content[:pos-1], tmp.Content[pos+2:]
			tmp.Content = append(b, v)
			tmp.Content = append(tmp.Content, e...)
			nextOp(pos-1, s) // step back one token
		}

	}

	log.Printf("[collapse] returns: %s", tlistStr("", tmp))

	return tmp, nil
}

func (d *LogoDriver) ParseExprRLCollapse(tokens *types.TokenList, subexpr bool) ([]*types.Token, error) {

	/*
		LIST 2 * 10 RANDOM 20 + 100
							  ^
							V   V
	*/
	var err error
	tokens, err = d.CollapseSimpleExpr(tokens)
	if err != nil {
		return nil, err
	}

	i := tokens.Size() - 1
	var ops = types.NewTokenList()
	var vals = types.NewTokenList()

	var processOps = func() error {
		for ops.Size() > 0 && vals.Size() >= 2 {
			t := ops.Shift()
			a := vals.Shift()
			b := vals.Shift()
			if b.Type == types.EXPRESSIONLIST {
				tmp, err := d.ParseExprRLCollapse(b.List, true)
				if err != nil {
					return err
				}
				b = tmp[0]
			}
			if a.Type == types.EXPRESSIONLIST {
				tmp, err := d.ParseExprRLCollapse(a.List, true)
				if err != nil {
					return err
				}
				a = tmp[0]
			}

			var isSame bool
			var diffDir int
			var aa, bb float64

			if t.Type == types.COMPARITOR || t.Type == types.ASSIGNMENT {
				isSame, diffDir = d.Compare(a, b)
				log.Printf("[rl] comparison result yields same %v, diffDir %d", isSame, diffDir)
			} else {
				aa = a.AsExtended()
				bb = b.AsExtended()
				log.Printf("[rl] expr eval %f %s %f", aa, t.Content, bb)
			}

			switch t.Content {
			case "<=":
				cv := "0"
				if isSame || diffDir == -1 {
					cv = "1"
				}
				vals.UnShift(types.NewToken(types.NUMBER, cv))
			case ">=":
				cv := "0"
				if isSame || diffDir == 1 {
					cv = "1"
				}
				vals.UnShift(types.NewToken(types.NUMBER, cv))
			case ">":
				cv := "0"
				if diffDir == 1 {
					cv = "1"
				}
				vals.UnShift(types.NewToken(types.NUMBER, cv))
			case "<":
				cv := "0"
				if diffDir == -1 {
					cv = "1"
				}
				vals.UnShift(types.NewToken(types.NUMBER, cv))
			case "=":
				cv := "0"
				if isSame {
					cv = "1"
				}
				vals.UnShift(types.NewToken(types.NUMBER, cv))
			case "+":
				vals.UnShift(types.NewToken(types.NUMBER, utils.FloatToStr(aa+bb)))
			case "-":
				vals.UnShift(types.NewToken(types.NUMBER, utils.FloatToStr(aa-bb)))
			case "*":
				vals.UnShift(types.NewToken(types.NUMBER, utils.FloatToStr(aa*bb)))
			case "/":
				if bb != 0 {
					vals.UnShift(types.NewToken(types.NUMBER, utils.FloatToStr(aa/bb)))
				} else {
					return errors.New("division by zero")
				}
			}
		}
		return nil
	}

	for i >= 0 {
		t := tokens.Get(i)
		log.Printf("[rl] next token at %d is %s(%s)", i, t.Type, t.AsString())
		switch t.Type {
		case types.COMPARITOR:
			ops.Push(t)
		case types.ASSIGNMENT:
			ops.Push(t)
		case types.OPERATOR:
			ops.Push(t)
		case types.FUNCTION:

			err = processOps()
			if err != nil {
				return nil, err
			}

			log.Printf("[rl] calling func %s with params: [%+v]", t.Content, vals.AsString())
			fun, ok := d.d.GetFunctions()[strings.ToLower(t.Content)]
			fun.SetEntity(d.ent)

			if !ok {
				return nil, errors.New("Unknown function: " + t.Content)
			}

			paramcount := fun.GetMaxParams()
			if paramcount > vals.Size() {
				paramcount = vals.Size()
			}
			if paramcount < vals.Size() && i == 0 && subexpr {
				paramcount = vals.Size()
			}

			log.Printf("[rl] calling function %s with %d params %+v", t.Content, paramcount, tlistStr("", vals.SubList(0, paramcount)))

			fun.SetAllowMoreParams(i == 0 && subexpr)
			e := fun.FunctionExecute(vals.SubList(0, paramcount))
			fun.SetAllowMoreParams(false)
			if e != nil {
				return nil, e
			}

			vals = vals.SubList(paramcount, vals.Size())
			log.Printf("[rl] stack after function now %s", tlistStr("", vals))
			fr := fun.GetStack().Pop()
			vals.UnShift(fr)
			log.Printf("[rl] putting result %s on stack", tokenStr("", fr))
			log.Printf("[rl] stack is now %s", tlistStr("", vals))

			//ops.Push(t)
		case types.DYNAMICFUNCTION, types.DYNAMICKEYWORD:

			dcmd, ok := d.GetProc(t.Content)

			if !ok {
				return nil, errors.New("I don't know how to " + t.Content)
			}

			log.Printf("calling proc %s", t.Content)

			paramcount := len(dcmd.Arguments)
			if paramcount > vals.Size() {
				paramcount = vals.Size()
			}
			if paramcount < vals.Size() && i == 0 && subexpr {
				paramcount = vals.Size()
			}

			log.Printf("[rl] calling proc %s with %d params %+v", t.Content, paramcount, tlistStr("", vals.SubList(0, paramcount)))

			ret, err := d.ExecProc(dcmd, vals.SubList(0, paramcount))
			if err != nil {
				return nil, err
			}

			d.Printf("ret value is %v", ret)
			vals = vals.SubList(paramcount, vals.Size())
			if ret != nil {
				vals.UnShift(ret)
				log.Printf("[rl] putting result %s on stack", ret.AsString())
			}

		case types.EXPRESSIONLIST:
			tmp, err := d.ParseExprRLCollapse(t.List, true)
			if err != nil {
				return nil, err
			}
			for _, tt := range tmp {
				vals.UnShift(tt)
			}
		case types.NUMBER:
			vals.UnShift(t)
		case types.VARIABLE:
			tt, _ := d.GetVar(t.Content)
			if tt == nil {
				return nil, errors.New("No such variable " + t.Content)
			}
			vals.UnShift(tt)
		case types.WORD:
			vals.UnShift(t)
		case types.LIST:
			vals.UnShift(t)
		}
		i--

		if ops.Size() > 0 && vals.Size() >= 2 {
			err = processOps()
			if err != nil {
				return nil, err
			}
		}
	}

	err = processOps()
	if err != nil {
		return nil, err
	}

	if vals.Size() == 0 {
		return nil, errors.New("expected result")
	}

	var out = []*types.Token{}
	for _, t := range vals.Content {
		out = append(out, t)
	}

	return out, nil

}

func tlistStr(in string, tl *types.TokenList) string {
	var out string
	for i, t := range tl.Content {
		if i > 0 {
			out += " "
		}
		out += tokenStr("", t)
	}
	out += " "
	return in + out
}

func tokenStr(in string, t *types.Token) string {

	var out string

	if t.Type == types.LIST {
		out += "["
		for i, tt := range t.List.Content {
			if i > 0 {
				out += " "
			}
			out = tokenStr(out, tt)
		}
		out += "]"
	} else {
		out += t.Content
	}

	return in + out

}
