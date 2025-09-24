package instrument

import (
	"net/http"
	"strings"

	"paleotronic.com/fmt"

	"paleotronic.com/octalyzer/bus"

	"paleotronic.com/core/hardware/apple2"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
	"paleotronic.com/octalyzer/backend"
)

func StartInstServer(port string) {
	http.HandleFunc("/", vmInfo)
	go http.ListenAndServe(port, nil)
}

func vmInfo(w http.ResponseWriter, r *http.Request) {

	lines := []string{
		"<html>",
		"<head>",
		"<title>", "MicroM8 State", "</title>",
		"</head>",
		"<body>",
		"<table border='1' width='100%'>",
		"<tr><th>VM#</th><th>Mode</th><th>Schema</th><th>State</th><th>Modes</th><th>Render</th><th>Clocked</th></tr>",
	}

	if backend.ProducerMain != nil {

		for _, ent := range backend.ProducerMain.GetInterpreterList() {
			if ent != nil {

				rendered := ent.GetMemoryMap().IntGetLayerState(ent.GetMemIndex()) != 0
				clocked := bus.IsClock()
				clockTime := bus.ClockTime

				state := ""
				dialect := ""
				modes := []string(nil)
				for _, l := range ent.GetHUDLayerSet() {
					if l != nil && l.GetActive() {
						modes = append(modes, l.String())
					}
				}
				for _, l := range ent.GetGFXLayerSet() {
					if l != nil && l.GetActive() {
						modes = append(modes, l.String())
					}
				}

				if mr, ok := ent.GetMemoryMap().InterpreterMappableAtAddress(ent.GetMemIndex(), 0xc000); ok {
					modes = append(modes, mr.(*apple2.Apple2IOChip).GetVidMode().String())
				}

				switch {
				case ent.GetState() == types.EXEC6502 || ent.GetState() == types.DIRECTEXEC6502:
					cpu := apple2helpers.GetCPU(ent)
					state = cpu.ToString()
					dialect = "MOS6502 emulation"
				case ent.GetState() == types.STOPPED:
					dialect = ent.GetDialect().GetTitle()
				case ent.GetState() == types.RUNNING:
					pc := ent.GetPC()
					l, ok := ent.GetCode().Get(pc.Line)
					state = fmt.Sprintf("Line %d, Statement %d", pc.Line, pc.Statement)
					dialect = ent.GetDialect().GetTitle()
					if ok {

						state += "<tt><table>"
						state += "<tr><td>" + fmt.Sprintf("%d ", pc.Line) + "</td>"
						for snum, s := range l {
							if snum != 0 {
								state += "<tr><td>:</td>"
							}
							if snum == pc.Statement {
								state += "<td><b>" + s.AsString() + "</b></td></tr>"
							} else {
								state += "<td>" + s.AsString() + "</td></tr>"
							}
						}
						state += "</table></tt>"
					}

				case ent.GetState() == types.DIRECTRUNNING:
					pc := ent.GetLPC()
					l, ok := ent.GetDirectAlgorithm().Get(pc.Line)
					state = fmt.Sprintf("Line %d, Statement %d", pc.Line, pc.Statement)
					dialect = ent.GetDialect().GetTitle()
					if ok {

						state += "<tt><table>"
						state += "<tr><td>" + fmt.Sprintf("%d ", pc.Line) + "</td>"
						for snum, s := range l {
							if snum != 0 {
								state += "<tr><td>:</td>"
							}
							if snum == pc.Statement {
								state += "<td><b>" + s.AsString() + "</b></td></tr>"
							} else {
								state += "<td>" + s.AsString() + "</td></tr>"
							}
						}
						state += "</table></tt>"
					}
				default:
					dialect = ent.GetDialect().GetTitle()
				}

				lines = append(lines,
					fmt.Sprintf("<tr><td>%d</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%v</td><td>%s</td></tr>",
						ent.GetMemIndex(),
						ent.GetState().String(),
						dialect,
						state,
						strings.Join(modes, "<br/>"),
						rendered,
						fmt.Sprintf("%v every %v", clocked, clockTime),
					),
				)
			}
		}

	}

	lines = append(lines,
		[]string{
			"</table>",
			"</body>",
			"</html>",
		}...)

	w.Write([]byte(strings.Join(lines, "\n")))
}
