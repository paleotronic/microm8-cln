package logo

var SharedVars *LogoVarTable

func init() {
	SharedVars = NewLogoVarTable()
}
