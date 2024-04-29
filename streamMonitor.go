package main

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type CameraStatus struct {
	Cam1Active   bool      // Status of the camera1
	Cam1Expiry   time.Time // Expiry timestamp of camera1
	Cam2Active   bool
	Cam2Expiry   time.Time
}

var (
	cameraMap     = make(map[string]CameraStatus)
	cameraMapLock sync.Mutex
)

func AddOrUpdateLiveStreamStatus(unitID string, camID string) string {
	// Define retVal
	var retVal = ""

	// Lock before accessing the cameraMap
	cameraMapLock.Lock()
	defer cameraMapLock.Unlock()

	// Check if the unitID entry exists in the map
	status, exists := cameraMap[unitID]
	if !exists {
		// If entry does not exist, initialize it with default values
		status = CameraStatus{
			Cam1Active: true,
			Cam1Expiry: time.Now().Add(1 * time.Minute),
			Cam2Active: false,
			Cam2Expiry: time.Now().Add(1 * time.Minute),
		}
	} else {
		// Update the respective camera status based on camID
		switch camID {
		case "cam1":
			if status.Cam1Active {
				// If cam1 is already active, return no_op
				retVal = "no_op"
			} else {
				// Activate cam1 and set expiry time
				status.Cam1Active = true
				status.Cam1Expiry = time.Now().Add(1 * time.Minute)

				// Check if cam2 is active, return add_chan if true, otherwise return add_stream
				if status.Cam2Active {
					retVal = "add_chan"
				} else {
					retVal = "add_stream"
				}
			}
		case "cam2":
			if status.Cam2Active {
				// If cam2 is already active, return no_op
				retVal = "no_op"
			} else {
				// Activate cam2 and set expiry time
				status.Cam2Active = true
				status.Cam2Expiry = time.Now().Add(1 * time.Minute)

				// Check if cam1 is active, return add_chan if true, otherwise return add_stream
				if status.Cam1Active {
					retVal = "add_chan"
				} else {
					retVal = "add_stream"
				}
			}
		default:
			retVal = "" // Invalid camID
		}
	}

	// Update the cameraMap entry with the modified status
	cameraMap[unitID] = status

	// Return the retVal
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
							if time.Now().After(status.Cam1Expiry) && status.Cam1Active {
									status.Cam1Active = false
									cameraMap[camID] = status
							}
							if time.Now().After(status.Cam2Expiry) && status.Cam2Active {
									status.Cam2Active = false
									cameraMap[camID] = status
							}
							// Remove entry if both cameras are inactive
							if !status.Cam1Active && !status.Cam2Active {
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