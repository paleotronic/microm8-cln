package interfaces

import (
	"errors"
	//"paleotronic.com/log"
	"strings"
)

type FunctionList map[string]Functioner

func NewFunctionList() FunctionList {
	return make(FunctionList)
}

func (this FunctionList) ContainsKey(s string) bool {
	_, ok := this[s]
	return ok
}

func (this FunctionList) Get(s string) Functioner {
	f, _ := this[s]

	return f
}

type NameSpaceFunctionList map[string]FunctionList

func NewNameSpaceFunctionList() NameSpaceFunctionList {
	return make(NameSpaceFunctionList)
}

func (this NameSpaceFunctionList) ContainsKey(s string) bool {
	_, ok := this[s]
	return ok
}

func (this NameSpaceFunctionList) Get(s string) FunctionList {
	f, _ := this[s]

	return f
}

func (this NameSpaceFunctionList) GetFunctionByNameContext( ns string, n string ) (Functioner, bool, string, error) {

	/*

	Could either get: -

	a.b.c{fgdfd}
	.c{rdffdf}

	*/

	fullname := n
	parts := strings.Split(fullname, ".")
	if len(parts) == 1 {
		return nil, false, "", errors.New("UNDEFINED FUNCTION")
	}
	funcname := parts[len(parts)-1]
	namespace := strings.Join( parts[0:len(parts)-1], "." )
//	log.Printf("*** GetFunctionByNameContext(): Looking for %s in namespace %s\n", funcname, namespace)

	//~ if namespace == "" {
		//~ namespace = ns
	//~ }

	//~ if namespace == "" {
		//~ i := 0
		//~ // search blind for first match
		//~ for tnamespace, fl := range this {
			//~ _, ok := fl[funcname]
			//~ if ok {
				//~ namespace = tnamespace
				//~ i++
			//~ }
		//~ }

		//~ if i > 1 {
			//~ return nil, false, "", errors.New("FN ABIGUITY ERROR")
		//~ }

		//~ if i == 0 {
			//~ return nil, false, "", errors.New("UNDEFINED FUNCTION")
		//~ }
	//~ }

//~ //	log.Printf("--> Resolving to use %s as namespace\n", namespace)

	//~ if !this.ContainsKey(namespace) {

		//~ i := 0
		//~ // search blind for first match
		//~ for tnamespace, fl := range this {
			//~ _, ok := fl[funcname]
			//~ if ok {
				//~ namespace = tnamespace
				//~ i++
			//~ }
		//~ }

		//~ if i > 1 {
			//~ return nil, false, "", errors.New("FN ABIGUITY ERROR")
		//~ }

		//~ if i == 0 {
			//~ return nil, false, "", errors.New("UNDEFINED FUNCTION")
		//~ }

		//~ if !this.ContainsKey(namespace) {
			//~ return nil, false, "", errors.New("UNDEFINED FUNCTION")
		//~ }
	//~ }

	nn := this.Get(namespace)

	f, x := nn[funcname]
	return f, x, namespace, nil

}
