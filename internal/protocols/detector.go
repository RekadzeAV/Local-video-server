package protocols

import (
	"time"

	"github.com/local-video-server/internal/models"
	"github.com/local-video-server/pkg/utils"
	"github.com/sirupsen/logrus"
)

// Detector - интерфейс для детекторов протоколов
type Detector interface {
	// Detect проверяет наличие протокола на устройстве
	Detect(ip string, port int, timeout time.Duration) (*models.Protocol, error)
	
	// GetName возвращает название протокола
	GetName() string
	
	// GetDefaultPort возвращает порт по умолчанию
	GetDefaultPort() int
}

// ProtocolDetector - координатор всех детекторов протоколов
type ProtocolDetector struct {
	detectors []Detector
	logger    *logrus.Logger
}

// NewProtocolDetector создает новый координатор детекторов
func NewProtocolDetector() *ProtocolDetector {
	logger := utils.GetLogger()
	
	return &ProtocolDetector{
		detectors: []Detector{
			NewRTSPDetector(),
			NewRTMPDetector(),
			NewHLSDetector(),
			NewMJPEGDetector(),
			NewDASHDetector(),
			NewWebRTCDetector(),
		},
		logger: logger,
	}
}

// DetectAll проверяет все протоколы на устройстве
func (pd *ProtocolDetector) DetectAll(ip string, timeout time.Duration) ([]models.Protocol, error) {
	var protocols []models.Protocol
	
	for _, detector := range pd.detectors {
		port := detector.GetDefaultPort()
		protocol, err := detector.Detect(ip, port, timeout)
		if err != nil {
			pd.logger.Debugf("Protocol %s not detected on %s:%d: %v", 
				detector.GetName(), ip, port, err)
			continue
		}
		
		if protocol != nil && protocol.Available {
			protocols = append(protocols, *protocol)
			pd.logger.Infof("Detected %s protocol on %s:%d", 
				detector.GetName(), ip, port)
		}
	}
	
	return protocols, nil
}

// DetectProtocol проверяет конкретный протокол
func (pd *ProtocolDetector) DetectProtocol(protocolName string, ip string, port int, timeout time.Duration) (*models.Protocol, error) {
	for _, detector := range pd.detectors {
		if detector.GetName() == protocolName {
			return detector.Detect(ip, port, timeout)
		}
	}
	
	return nil, nil
}
