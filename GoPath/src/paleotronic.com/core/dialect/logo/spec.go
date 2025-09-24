package logo

import (
	"paleotronic.com/core/interfaces"
)

var logoSpec = `
pattern DIGIT			[0-9]
pattern HEXDIGIT        [A-Fa-f]|{DIGIT}
pattern ALPHA  			[A-Za-z]
pattern JOIN            [._?]
pattern IDENTIFIER		{ALPHA}({DIGIT}|{ALPHA}|{JOIN})*
pattern WHITESPACE  	[ \t\r\n]+
pattern INTEGER     	[-]?{DIGIT}+
pattern FLOAT           [-]?{DIGIT}*[.]{DIGIT}+([eE][+-]?{DIGIT}+)?
pattern STRING 	        [a-zA-Z0-9_/."'!><]+   
pattern HEXNUM          [$]{HEXDIGIT}+
pattern BLOCKSTART      \[
pattern BLOCKEND        \]
pattern LPAREN      	\(
pattern RPAREN        	\)
pattern MULDIV			[*/]
pattern ADDSUB          [-+]
pattern BOOLOP          and|or|not
pattern COMPARE         <|>|<=|>=|==|!=|<>

token escapedchar       \\.
token whitespace 		{WHITESPACE}
token identifier 		{IDENTIFIER}
token blockbeg         	\[
token blockend         	\]
token listbeg          	\(
token listend          	\)
token string 	  		["]{STRING}
token int               {INTEGER}
token hex               {HEXNUM}
token float             {FLOAT}
token varref 	        [:]{IDENTIFIER}
token muldiv            {MULDIV}
token addsub            {ADDSUB}
token .comment          [;][;].*
token comparison        {COMPARE}
token assign            [=]
token symbols           [a-zA-Z0-9~!@#$%^&{}:;.,?'"_<>]+
`

func GetSpec(dia interfaces.Dialecter) string {
	return logoSpec
}
