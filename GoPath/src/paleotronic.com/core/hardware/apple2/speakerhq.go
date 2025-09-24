package apple2

import "paleotronic.com/core/settings"

// WriteSampleHW outputs one HQ sample to the buffer
func (s *AppleSpeaker) WriteSampleHQ() {
	s.SampleCounter++

	if settings.AudioUsesLeapTicks {
		if s.SampleCounter%s.LeapTicksEvery == 0 {
			s.AdjustTicksPerSample = s.TicksPerSample + s.LeapTickSign
		} else {
			s.AdjustTicksPerSample = s.TicksPerSample
		}
	}

	//clip := int(float64(*s.HQLevelMax) * s.divfactor)
	var v float32
	if settings.MuteCPU {
		v = 0
	} else {
		v = float32(s.HQLevel) * 0.04 //* (1 / float32(clip))
	}

	s.HQBuffers[s.CurrentBuffer][s.SampleCount] = v
	if v == 0 {
		s.emptyCount++
	}

	// if s.HQBuffers[s.CurrentBuffer][s.SampleCount] > 0 {
	// 	log2.Printf("%f", s.HQBuffers[s.CurrentBuffer][s.SampleCount])
	// }

	s.HQLevel /= 2

	s.SampleCount++
	if s.SampleCount >= s.BufferSize {

		index := s.e.GetMemIndex()

		// output sink
		if s.uid == "CAS" && settings.RecordC020[index] {
			settings.RecordC020Buffer[index] = append(settings.RecordC020Buffer[index], s.HQBuffers[s.CurrentBuffer]...)
			settings.RecordC020Rate[index] = s.SampleRate
		}

		if s.uid != "CAS" || s.emptyCount != len(s.HQBuffers[s.CurrentBuffer]) {
			s.outputFuncNB(s.channel, s.HQBuffers[s.CurrentBuffer], false, s.SampleRate)
		}
		if s.e.IsRecordingVideo() && (s.emptyCount != len(s.HQBuffers[s.CurrentBuffer])) {
			s.e.GetMemoryMap().RecordSendAudioPackedF(s.e.GetMemIndex(), s.channel, s.SampleCount, s.HQBuffers[s.CurrentBuffer], s.SampleRate, false)
		}

		s.SampleCount = 0
		s.CurrentBuffer = (s.CurrentBuffer + 1) % MAX_SPEAKER_BUFFERS
		s.emptyCount = 0

	}
}
