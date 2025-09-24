package logo

import "paleotronic.com/core/types"

func (d *LogoDriver) CrunchLists(tokens *types.TokenList) *types.TokenList {
	var out = types.NewTokenList()
	var bc int
	var subexpr *types.TokenList
	var closure string
	for i, _ := range tokens.Content {
		t := tokens.Content[i].Copy()
		//d.Printf("Got token: %s, %s", t.Type, t.Content)
		switch t.Type {
		case types.OBRACKET:
			if bc > 0 {
				if closure == ")" && t.Content == "(" {
					bc++
				} else if closure == "]" && t.Content == "[" {
					bc++
				}
				subexpr.Push(t)
			} else {
				// start subexpr without the opening bracket
				switch t.Content {
				case "(":
					closure = ")"
				case "[":
					closure = "]"
				}
				subexpr = types.NewTokenList()
				bc = 1
			}
		case types.CBRACKET:
			if bc > 0 {
				if closure == t.Content {
					bc--
					if bc == 0 {
						// we've closed the loop
						//d.Printf("raw subexpr is length %d", subexpr.Size())
						newexpr := d.CrunchLists(subexpr)
						//d.Printf("raw newexpr is length %d", newexpr.Size())
						switch closure {
						case "]":
							//d.Printf("processed rec BLOCK of length %d", newexpr.Size())
							tt := types.NewToken(types.LIST, "")
							tt.List = newexpr
							out.Push(tt)
						case ")":
							if newexpr.Size() > 0 && (newexpr.LPeek().Type == types.KEYWORD || newexpr.LPeek().Type == types.DYNAMICKEYWORD) {
								//d.Printf("processed rec COMMAND LIST of length %d", newexpr.Size())
								tt := types.NewToken(types.COMMANDLIST, "")
								tt.List = newexpr
								out.Push(tt)
							} else {
								//d.Printf("processed rec EXPR LIST of length %d", newexpr.Size())
								tt := types.NewToken(types.EXPRESSIONLIST, "")
								tt.List = newexpr
								out.Push(tt)
							}
						}
					} else {
						// still collecting
						subexpr.Push(t)
					}
				} else {
					subexpr.Push(t)
				}
			}
		default:
			if bc > 0 {
				subexpr.Push(t)
				//d.Printf("pushing to subexpr")
			} else {
				//d.Printf("pushing to output")
				out.Push(t)
			}
		}
	}
	//d.Printf("returning list containing %d items", out.Size())
	return out
}
