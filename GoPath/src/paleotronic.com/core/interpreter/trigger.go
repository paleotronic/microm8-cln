package interpreter

import "paleotronic.com/core/types"
import "paleotronic.com/core/interfaces"
import "paleotronic.com/fmt"

type Trigger struct {
	Slot       int
	Condition  *types.TokenList
	Repeat     bool // autoreset = false by default
	Done       bool
	TargetLine *types.CodeRef // This effects a gosub to this location...  do something and return
}

func (t *Trigger) IsTrue(ent interfaces.Interpretable) bool {

	if t.Done {
		return false
	}

	rtok := ent.ParseTokensForResult(*t.Condition)
	if rtok.AsInteger() != 0 {
		// triggered
		if !t.Repeat {
			t.Done = true
		}
		return true
	}

	return false
}

func (t *Trigger) Reset() {
	t.Done = false
}

func (t *Trigger) Disable() {
	t.Done = true
}

/* Trigger table */
/*
 * OCTALYZER_TRIGGERS_START, OCTALYZER_TRIGGERS_END
 *
 * These vectors control the trigger table position.
 */

type TriggerTable struct {
	List         []*Trigger
	InTrigger    bool // set to true if we branch, cleared if we return...
	TriggerLevel int  // Stack level of last trigger
	Int          *Interpreter
}

func NewTriggerTable(ent *Interpreter) *TriggerTable {

	tt := &TriggerTable{
		List:         make([]*Trigger, 0),
		InTrigger:    false,
		TriggerLevel: -1,
		Int:          ent,
	}

	return tt

}

func (tt *TriggerTable) Pop() {
	tt.Return()
}

func (tt *TriggerTable) Return() {
	if tt.InTrigger {
		if tt.Int.Stack.Size() <= tt.TriggerLevel {
			tt.InTrigger = false
		}
	}
}

func (tt *TriggerTable) Test() bool {

	if tt.InTrigger {
		return false
	}

	remove := -1
	triggered := false

	for ii, t := range tt.List {

		if triggered {
			continue
		}

		// Which slot is this trigger for...
		sb, se := t.Slot, t.Slot
		if sb == -1 {
			sb = 0
			se = 7
		}

		for i := sb; i <= se; i++ {

			if triggered {
				continue
			}

			e := tt.Int.GetProducer().GetInterpreter(i)
			if t.IsTrue(e) {
				// we've disabled the trigger, but we should remove it...
				remove = ii
				triggered = true
				cr := *t.TargetLine
				caller := tt.Int
				tt.TriggerLevel = caller.GetStack().Size()
				tt.InTrigger = true
				a := caller.GetCode()

				caller.CallTrigger(cr, a, caller.GetState(), false, caller.GetVarPrefix(), *caller.GetTokenStack(), caller.GetDialect())
				fmt.Printf("!!! Triggered: [%d][%s] -> %d\n", i, caller.TokenListAsString(*t.Condition), cr.Line)
			}
		}

	}

	if triggered {
		a := tt.List[:remove]
		b := tt.List[remove+1:]
		tt.List = append(a, b...)
	}

	return triggered

}

func (tt *TriggerTable) Add(target int, condition *types.TokenList, line int) {

	l := types.NewCodeRef()
	l.Line = line

	t := &Trigger{

		Condition:  condition,
		Slot:       target,
		Repeat:     false,
		Done:       false,
		TargetLine: l,
	}

	tt.List = append(tt.List, t)
	fmt.Printf("+++ Set-Trigger: [%d][%s] -> %d\n", t.Slot, tt.Int.TokenListAsString(*t.Condition), line)

}

func (tt *TriggerTable) Empty() {
	tt.List = make([]*Trigger, 0)
	tt.InTrigger = false
	tt.TriggerLevel = tt.Int.Stack.Size()
}
