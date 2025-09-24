package syncmanager

import (
	"strings"
	"paleotronic.com/api"
	"paleotronic.com/core/types"
	"paleotronic.com/log"
)

var Sync VariableSyncher

/* Handle variable sync */
type VariableSyncher struct {
	conn *s8webclient.Client
	synckey string
}

func NewVariableSyncher( conn *s8webclient.Client ) *VariableSyncher {
	this := &VariableSyncher{ conn: conn }	
	return this
}

/* SYK <md5>:<varname>:<stacked> */
func (vs *VariableSyncher) SetSyncKey( md5 string ) error {
	
	if vs.conn == nil {
		vs.conn = s8webclient.CONN
	}
	
	log.Printf("<--> Setting sync key to [%s]\n", md5)
	
	vs.synckey = md5
	
//	if vs.conn == nil {
//		log.Println("vs.conn is not defined!")
//		return nil
//	}
	
//	msg, err := vs.conn.SendAndWait( "SYK", []byte(vs.conn.Session+":"+md5), []string{"SYO"} )
	
//	log.Println(msg)
	
//	return err

	return nil
	
}

/* VUP <md5>:<varname>:<value> */
func (vs *VariableSyncher) PublishUpdate( varname string, varvalue types.VarValue ) error {
	
	v := varvalue.Value
	if varvalue.Pending() > 0 {
		if varvalue.Content[varvalue.Pending()-1].Published == false {
			v = varvalue.Content[varvalue.Pending()-1].Value
		} else {
			return nil // no publish, already done
		}
	}
	
	msg, err := vs.conn.SendAndWait( "VUP", []byte(vs.conn.Session+":"+vs.synckey+":"+varname+":"+v), []string{"VUO"} )
	
	log.Println(msg)
	
	return err
	
}

func (vs *VariableSyncher) PullUpdate( varname string ) (types.VarValue, error) {
	
	msg, err := vs.conn.SendAndWait( "VPU", []byte(vs.conn.Session+":"+vs.synckey+":"+varname), []string{"VPO"} )
	
	log.Println(msg)
	
	// decode  
	var v types.VarValue
	
	// format
	// <value>(0 byte)<value>(0 byte)
	if err != nil {
		return v, err
	}
	
	parts := strings.Split( string(msg.Payload), string(rune(0)) )
	
	for i, vv := range parts {
		if i == 0 {
			v.Assign(vv, false)
		} else {
			v.Assign(vv, true)
		}
	}
	
	return v, nil
	
}

