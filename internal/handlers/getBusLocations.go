package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/transitIOM/projectMercury/api"
	"github.com/transitIOM/projectMercury/internal/tools"
)

func GetBusLocations(w http.ResponseWriter, r *http.Request) {
	log.Debug("Handling getBusLocations request")
	tools.BusLocations.Mutex.RLock()
	defer tools.BusLocations.Mutex.RUnlock()
	busLocations := tools.BusLocations.Data
	nullifyObsoleteValues(&busLocations)

	busLocationsBytes, err := json.Marshal(busLocations)
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(busLocationsBytes)
}

func nullifyObsoleteValues(busLocations *[]tools.BusLocation) {
	for i := range *busLocations {
		(*busLocations)[i].DriverNumber = ""
		(*busLocations)[i].Timestamp = time.Time{}
		(*busLocations)[i].Unknown1 = 0
		(*busLocations)[i].Unknown2 = ""
	}
}
