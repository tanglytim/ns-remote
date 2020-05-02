package server

import (
	"errors"
	"log"
	"math/rand"

	"github.com/pion/webrtc/v2"
)

// MediaStreamer is..
type MediaStreamer struct {
	peerConnection *webrtc.PeerConnection
	VideoTrack     *webrtc.Track
	AudioTrack     *webrtc.Track
}

// Setup is ..
func (m *MediaStreamer) Setup(offer webrtc.SessionDescription) {
	// WebRTC setup
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}
	mediaEngine := webrtc.MediaEngine{}
	mediaEngine.PopulateFromSDP(offer)

	api := webrtc.NewAPI(webrtc.WithMediaEngine(mediaEngine))

	var err error
	m.peerConnection, err = api.NewPeerConnection(config)
	if err != nil {
		panic(err)
	}

	m.peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		log.Printf("Connection State has changed %s \n", connectionState.String())
	})

	// Create a video track
	videoCodec, err := findCodecOfType(mediaEngine, webrtc.RTPCodecTypeVideo, webrtc.H264)
	if err != nil {
		panic(err)
	}

	m.VideoTrack, err = m.peerConnection.NewTrack(videoCodec.PayloadType, rand.Uint32(), "video", "video")
	if err != nil {
		panic(err)
	}
	_, err = m.peerConnection.AddTransceiverFromTrack(m.VideoTrack,
		webrtc.RtpTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionSendonly,
		},
	)
	if err != nil {
		panic(err)
	}

	// Create a audio track
	audioCodec, err := findCodecOfType(mediaEngine, webrtc.RTPCodecTypeAudio, webrtc.Opus)
	if err != nil {
		panic(err)
	}
	m.AudioTrack, err = m.peerConnection.NewTrack(audioCodec.PayloadType, rand.Uint32(), "audio", "audio")
	if err != nil {
		panic(err)
	}
	_, err = m.peerConnection.AddTransceiverFromTrack(m.AudioTrack,
		webrtc.RtpTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionSendonly,
		},
	)
	if err != nil {
		panic(err)
	}
}

// CreateAnswerFromOffer is..
func (m *MediaStreamer) CreateAnswerFromOffer(offer webrtc.SessionDescription) webrtc.SessionDescription {
	// Set the remote SessionDescription
	err := m.peerConnection.SetRemoteDescription(offer)
	if err != nil {
		panic(err)
	}

	// Create an answer
	answer, err := m.peerConnection.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}

	// Sets the LocalDescription, and starts our UDP listeners
	err = m.peerConnection.SetLocalDescription(answer)
	if err != nil {
		panic(err)
	}

	return answer
}

// findCodecOfType is..
func findCodecOfType(mediaEngine webrtc.MediaEngine, kind webrtc.RTPCodecType, codecName string) (*webrtc.RTPCodec, error) {
	codecs := mediaEngine.GetCodecsByKind(kind)
	for _, codec := range codecs {
		if codec.Name == codecName {
			return codec, nil
		}
	}
	return nil, errors.New("No codec of type " + codecName + " found")
}
