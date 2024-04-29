package main

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type CameraStatus struct {
	Active   bool      // Status of the camera1
	Expiry   time.Time // Expiry timestamp of camera1
}

var (
	cameraMap     = make(map[string]CameraStatus)
	cameraMapLock sync.Mutex
)

func AddOrUpdateLiveStreamStatus(camID string) string {
	// Initialize return value
	retVal := "updated"

	// Lock before accessing the cameraMap
	cameraMapLock.Lock()
	defer cameraMapLock.Unlock()

	// Check if the unitID entry exists in the map
	status, exists := cameraMap[camID]
	if !exists {
		// If entry does not exist, initialize it with default values
		status = CameraStatus{
			Active: true,
			Expiry: time.Now().Add(1 * time.Minute),
		}

		// Add code to start streaming
		retVal = "add"
	} else {
		status.Expiry = time.Now().Add(1 * time.Minute)
	}

	// Update the cameraMap entry with the modified status
	cameraMap[camID] = status

	// Return the status
	return retVal
}

func monitorCameraStatus(stopCh <-chan struct{}) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
			select {
			case <-ticker.C:
					cameraMapLock.Lock() // Lock outside the loop
					for camID, status := range cameraMap {
							// Delete the stream and entry from map, if expired
							if time.Now().After(status.Expiry) {
								StreamDelete(camID)
								delete(cameraMap, camID)
							}
					}
					cameraMapLock.Unlock() // Unlock outside the loop

        case <-stopCh:
					log.WithFields(logrus.Fields{
						"module": "main",
						"func":   "monitorCameraStatus",
					}).Info("Server stop working by signal")
					return // Stop the function when receiving the stop signal
		}
	}
}