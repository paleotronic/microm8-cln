package logo

import (
	"errors"
	"math"
	"strings"
	"time"

	tomb "gopkg.in/tomb.v2"
	"paleotronic.com/log"

	"github.com/go-gl/mathgl/mgl64"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/core/types/glmath"
	"paleotronic.com/utils"
)

func (d *LogoDriver) Track() {
	index := d.ent.GetMemIndex()
	mm := d.ent.GetMemoryMap()
	cindex := 1
	control := types.NewOrbitController(mm, index, cindex)

	mpos := mgl64.Vec3{0, 0, 0}
	vect := apple2helpers.VECTOR(d.ent)
	vl := apple2helpers.GETGFX(d.ent, "VCTR")
	t := vect.Turtle()
	tpos := mgl64.Vec3{
		t.Position[0],
		t.Position[1],
		t.Position[2],
	}
	glh := float64(types.CHEIGHT)
	glw := float64(types.CHEIGHT) * control.GetAspect()
	width := float64(vl.GetWidth())
	height := float64(vl.GetHeight())
	apos := utils.TurtleCoordinatesToGL(glw, glh, width, height, mpos, tpos)

	campos := control.GetPosition()

	// this is the distance between the camera and the turtle in GL space
	turtlepos := glmath.Vec3(apos[0], apos[1], apos[2])
	//dist := campos.Sub(turtlepos).Len()
	//turtleupdir := glmath.Vec3(t.UpDir[0], t.UpDir[1], t.UpDir[2])

	// tyaw := t.Heading
	// tpitch := t.Pitch
	// troll := t.Roll

	// new position based on up dir
	//newpos := turtlepos.Add(turtleupdir.Normalize().MulF(dist))
	control.LookAtWithPosition(campos, turtlepos)
	//control.SetLookAtV(turtlepos)
	control.Update()
	// control.SetRotation(glmath.Vec3(tpitch, troll, tyaw))
	// control.Update()
}

func (d *LogoDriver) TrackBehind() {
	index := d.ent.GetMemIndex()
	mm := d.ent.GetMemoryMap()
	cindex := 1
	control := types.NewOrbitController(mm, index, cindex)

	mpos := mgl64.Vec3{0, 0, 0}
	vect := apple2helpers.VECTOR(d.ent)
	vl := apple2helpers.GETGFX(d.ent, "VCTR")
	t := vect.Turtle()
	tpos := mgl64.Vec3{
		t.Position[0],
		t.Position[1],
		t.Position[2],
	}
	glh := float64(types.CHEIGHT)
	glw := float64(types.CHEIGHT) * control.GetAspect()
	width := float64(vl.GetWidth())
	height := float64(vl.GetHeight())
	apos := utils.TurtleCoordinatesToGL(glw, glh, width, height, mpos, tpos)

	campos := control.GetPosition()

	// this is the distance between the camera and the turtle in GL space
	turtlepos := glmath.Vec3(apos[0], apos[1], apos[2])
	dist := campos.Sub(turtlepos).Len()
	turtleupdir := glmath.Vec3(t.ViewDir[0], t.ViewDir[1], t.ViewDir[2])

	// tyaw := t.Heading
	// tpitch := t.Pitch
	// troll := t.Roll

	// new position based on up dir
	newpos := turtlepos.Add(turtleupdir.Normalize().MulF(-dist))
	control.LookAtWithPosition(newpos, turtlepos)
	control.Update()
	// control.SetRotation(glmath.Vec3(tpitch, troll, tyaw))
	// control.Update()
}

func (d *LogoDriver) TrackTurtle() {

	if !d.Tracking.FollowPosition {
		return
	}

	if d.Tracking.FollowBehind {
		d.TrackBehind()
	} else {
		d.Track()
	}

}

func (d *LogoDriver) SplitStmts(list *types.TokenList) ([]*types.TokenList, error) {
	//d.d.SplitOnTokenTypes()
	return nil, nil
}

func (d *LogoDriver) ExecuteStmt(stmt *types.TokenList) error {

	if !settings.LogoFastDraw[d.ent.GetMemIndex()] {
		for time.Since(d.lastExec) < d.InstDelay {
			time.Sleep(500 * time.Microsecond)
		}
	}

	if stmt == nil || stmt.Size() == 0 {
		return nil
	}

	log.Printf("[exec-stmt] executing statement %s", d.ent.TokenListAsString(*stmt))

	var tok = stmt.Get(0)
	var params = stmt.SubList(1, stmt.Size())

	defer d.TrackTurtle()

	switch tok.Type {
	case types.COMMANDLIST:
		return d.ExecuteStmt(tok.List)
	case types.IDENTIFIER:
		name := strings.ToLower(tok.Content)
		return errors.New("i don't know how to " + name)
	case types.KEYWORD:
		name := strings.ToLower(tok.Content)
		cmd, ok := d.d.GetCommands()[name]
		if !ok {
			return errors.New("i don't know how to " + name)
		}
		cmd.SetD(d.d)
		_, err := cmd.Execute(nil, d.ent, *params, nil, types.CodeRef{})
		return err
	case types.DYNAMICKEYWORD:
		name := strings.ToLower(tok.Content)
		proc, ok := d.Procs[name]
		if !ok {
			return errors.New("i don't know how to " + name)
		}
		return d.CreateProcScope(proc, d.PrepareParams(params))
	case types.FUNCTION, types.DYNAMICFUNCTION:
		params.UnShift(tok)
		result, err := d.ParseExprRLCollapse(params, true)
		if err != nil {
			return err
		}
		return errors.New("i don't know what to do with " + result[0].Content)
	default:
		return errors.New("i don't know what to do with " + tok.Content)
	}

	return nil
}

func (d *LogoDriver) PrepareParams(list *types.TokenList) *types.TokenList {
	r, _ := d.ParseExprRLCollapse(list, false)
	//log.Printf("params after parse: %s", d.DumpObjectStruct(r[0], false, ""))
	//if r[0].Type != types.LIST {
	// if r.Type == types.LIST {
	// 	return r.List
	// }
	tl := types.NewTokenList()
	// tl.Push(r)
	for _, t := range r {
		//log.Printf("prepare-params: param %d: %s", i, d.DumpObjectStruct(t, false, ""))
		tl.Push(t)
	}
	return tl
	//}
	//return r[0].List
}

func (d *LogoDriver) CanExec() bool {
	if d.S == nil {
		return false
	}
	if d.S.ProcRef == nil {
		return false
	}
	if d.S.ProcLine == -1 || d.S.ProcLine >= len(d.S.ProcRef.Lines) {
		return false
	}
	if d.S.ProcStmt == -1 || d.S.ProcStmt >= len(d.S.ProcRef.Lines[d.S.ProcLine].Statements) {
		return false
	}
	return true
}

func (d *LogoDriver) Printf(format string, args ...interface{}) {
	return
	pad := ""
	for i := 0; i < d.Stack.Size(); i++ {
		pad += "  "
	}
	log.Printf(pad+format, args...)
}

func (d *LogoDriver) Exec() (moved bool, err error) {

	if !d.CanExec() {
		return
	}

	code := d.S.ProcRef.Lines[d.S.ProcLine].Statements[d.S.ProcStmt]
	d.Printf("position: %s, line: %d, stmt: %d", d.S.ProcRef.Name, d.S.ProcLine, d.S.ProcStmt)

	var cline, cstmt = d.S.ProcLine, d.S.ProcStmt
	var cname = d.S.ProcRef.Name

	if code != nil {
		// exec statement
		err := d.ExecuteStmt(code)
		if err != nil {
			d.ErrScope = d.S
			// TODO: Write graceful stack collapse
			//d.CollapseStack()
			d.State = lsHalt
			d.Printf("got err = %v", err)
			return false, err
		}
	}

	if d.S == nil {
		return true, nil
	}

	moved = cname != d.S.ProcRef.Name || cline != d.S.ProcLine || cstmt != d.S.ProcStmt

	return moved, nil
}

func (d *LogoDriver) ReturnFromProc(r *types.Token) {

	ps := d.GetProcScope()
	for d.S != ps {
		d.Printf("return-from-proc: collapsing scope %s", d.S.ProcRef.Name)
		d.DestroyScope()
	}

	d.S.ReturnValue = r
	d.S.ProcLine = 99999999
	d.S.ProcStmt = 99999999
	d.S.StmtIterCount = 0
	if r != nil {
		d.Printf("Set %v as the return", r.Content)
	}
}

func (d *LogoDriver) ExecProc(p *LogoProc, args *types.TokenList) (*types.Token, error) {
	frame := d.Stack.Size()
	err := d.CreateProcScope(p, args)
	//log2.Printf("create-proc-scope returns %v", err)
	if err != nil {
		return nil, err
	}
	return d.ExecTillStackLevelReached(frame)
}

func (d *LogoDriver) ExecTillReturn() (*types.Token, error) {
	return d.ExecTillStackLevelReached(0)
}

func (d *LogoDriver) ProcessWhileIter(nline, nstmt int) (int, int, error) {
	res, err := d.ParseExprRLCollapse(d.S.StmtIterExpr.Copy(), false)
	if err != nil {
		return nline, nstmt, err
	}
	if len(res) > 0 && res[0].AsInteger() != 0 {
		///log2.Printf("while expr=%s, res=%s", tlistStr("", d.S.StmtIterExpr), tokenStr("", res[0]))
		nline, nstmt = 0, 0 // loop if condition is true
	}
	return nline, nstmt, nil
}

func (d *LogoDriver) ProcessForIter(nline, nstmt int) (int, int, error) {

	if d.S.StmtIterVar == nil || d.S.StmtIterDir == nil || d.S.StmtIterTarget == nil {
		return nline, nstmt, errors.New("ITER EXPECTS VAR, TARGET AND DIR")
	}

	v, scope := d.GetVar(*d.S.StmtIterVar)
	if v == nil {
		v = types.NewToken(types.NUMBER, "0")
	}
	if scope == nil {
		scope = d.S // current scope
	}
	var loop = math.Abs(v.AsExtended()-*d.S.StmtIterTarget) >= d.S.StmtIterStep

	// if *d.S.StmtIterDir < 0 && v.AsExtended() > *d.S.StmtIterTarget {
	// 	loop = true
	// } else if *d.S.StmtIterDir > 0 && v.AsExtended() < *d.S.StmtIterTarget {
	// 	loop = true
	// }
	if loop {
		v.Content = utils.FloatToStr(v.AsExtended() + *d.S.StmtIterDir*math.Abs(d.S.StmtIterStep))
		scope.Vars.Set(*d.S.StmtIterVar, v)
		///log2.Printf("while expr=%s, res=%s", tlistStr("", d.S.StmtIterExpr), tokenStr("", res[0]))
		nline, nstmt = 0, 0 // loop if condition is true
	}
	return nline, nstmt, nil
}

func (d *LogoDriver) ExecTillStackLevelReached(targetStack int) (*types.Token, error) {
	if !d.CanExec() {
		return nil, nil
	}

	for d.Stack.Size() > targetStack && !(d.BreakFunc != nil && d.BreakFunc()) && !(d.HasReboot()) {
		d.CheckPaused()
		moved, err := d.Exec()
		if err != nil {
			return nil, err
		}
		if d.hasResumed {
			d.hasResumed = false
			moved = false // this will cause positon to advance to after the pause instruction was issued
		}
		d.lastExec = time.Now()
		if !moved {
			nline, nstmt := d.S.GetNext()

			if nline == -1 && nstmt == -1 && d.S.StmtIterExpr != nil {
				var err error
				nline, nstmt, err = d.ProcessWhileIter(nline, nstmt)
				if err != nil {
					return d.LastReturn, err
				}
			} else if nline == -1 && nstmt == -1 && d.S.StmtIterVar != nil {
				var err error
				nline, nstmt, err = d.ProcessForIter(nline, nstmt)
				if err != nil {
					return d.LastReturn, err
				}
			}

			d.S.ProcLine = nline
			d.S.ProcStmt = nstmt
			for nline == -1 && d.Stack.Size() > targetStack {
				s, _ := d.DestroyScope()
				if s != nil {
					d.LastReturn = s.ReturnValue
				}
				if d.S != nil {
					nline, nstmt = d.S.GetNext()

					if nline == -1 && nstmt == -1 && d.S.StmtIterExpr != nil {
						var err error
						nline, nstmt, err = d.ProcessWhileIter(nline, nstmt)
						if err != nil {
							return d.LastReturn, err
						}
					} else if nline == -1 && nstmt == -1 && d.S.StmtIterVar != nil {
						var err error
						nline, nstmt, err = d.ProcessForIter(nline, nstmt)
						if err != nil {
							return d.LastReturn, err
						}
					}

					d.S.ProcLine = nline
					d.S.ProcStmt = nstmt
				}
			}
		} else {
			if d.S != nil {
				d.LastReturn = d.S.ReturnValue
			} else {
				d.LastReturn = nil
			}
		}
		d.CheckCommands() // handle commands
	}
	return d.LastReturn, nil
}

func (d *LogoDriver) ExecTillStackLevelReachedWithTomb(targetStack int, t *tomb.Tomb) (*types.Token, error) {
	if !d.CanExec() {
		return nil, nil
	}

	var dying bool

	for !dying && d.Stack.Size() > targetStack && !(d.BreakFunc != nil && d.BreakFunc()) && !(d.HasReboot()) {

		select {
		case <-t.Dying():
			//log2.Printf("breaking out")
			dying = true
			continue
		default:
		}

		d.CheckPaused()
		moved, err := d.Exec()
		if err != nil {
			return nil, err
		}
		if d.hasResumed {
			d.hasResumed = false
			moved = false // this will cause positon to advance to after the pause instruction was issued
		}
		d.lastExec = time.Now()
		if !moved {
			nline, nstmt := d.S.GetNext()
			if nline == -1 && nstmt == -1 && d.S.StmtIterExpr != nil {
				var err error
				nline, nstmt, err = d.ProcessWhileIter(nline, nstmt)
				if err != nil {
					return d.LastReturn, err
				}
			} else if nline == -1 && nstmt == -1 && d.S.StmtIterVar != nil {
				var err error
				nline, nstmt, err = d.ProcessForIter(nline, nstmt)
				if err != nil {
					return d.LastReturn, err
				}
			}
			d.S.ProcLine = nline
			d.S.ProcStmt = nstmt
			for nline == -1 && d.Stack.Size() > targetStack {
				s, _ := d.DestroyScope()
				if s != nil {
					d.LastReturn = s.ReturnValue
				}
				if d.S != nil {
					nline, nstmt = d.S.GetNext()

					if nline == -1 && nstmt == -1 && d.S.StmtIterExpr != nil {
						var err error
						nline, nstmt, err = d.ProcessWhileIter(nline, nstmt)
						if err != nil {
							return d.LastReturn, err
						}
					} else if nline == -1 && nstmt == -1 && d.S.StmtIterVar != nil {
						var err error
						nline, nstmt, err = d.ProcessForIter(nline, nstmt)
						if err != nil {
							return d.LastReturn, err
						}
					}

					d.S.ProcLine = nline
					d.S.ProcStmt = nstmt
				}
			}
		} else {
			if d.S != nil {
				d.LastReturn = d.S.ReturnValue
			} else {
				d.LastReturn = nil
			}
		}
		d.CheckCommands() // handle commands
	}
	return d.LastReturn, nil
}
