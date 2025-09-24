package types

type CameraCommand int
const (
	CC_None 		CameraCommand = iota
	CC_AbsPos
	CC_RelPos
	CC_RotX
	CC_RotY
	CC_RotZ
	CC_RotateAxis
	CC_Orbit
	CC_ResetAngle
	CC_ResetPos
	CC_ResetLookAt
	CC_ResetAll
	CC_PivotLock
	CC_PivotUnlock
	CC_LookAt
	CC_Zoom
	CC_Dolly
	CC_JSON
	CC_GetJSON
	CC_GetJSONR
	CC_Shake
	CC_GetView  // returns string with ex,ey,ez,lx,ly,lz,ux,uy,uz
	CC_SetView
)

type RestalgiaCommand uint 
const (
	  RS_None		  RestalgiaCommand = iota
      RS_Instrument	  
      RS_PlayNotes
      RS_PlaySong    
      RS_StopSong
      RS_PauseSong
      RS_ResumeSong
      RS_Sound
)
