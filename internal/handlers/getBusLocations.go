package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/transitIOM/projectMercury/api"
	"github.com/transitIOM/projectMercury/internal/tools"
)

// GetBusLocations godoc
// @Summary      Get current bus locations
// @Description  Retrieves real-time GPS locations for all active buses.
// @Tags         locations
// @Produce      json
// @Success      200  {object}  api.GetBusLocationsResponse
// @Failure      500  {object}  api.Error
// @Router       /locations/ [get]
func GetBusLocations(w http.ResponseWriter, r *http.Request) {
	log.Debug("Handling getBusLocations request")
	tools.BusLocations.Mutex.RLock()
	memBusLocations := tools.BusLocations.Data
	tools.BusLocations.Mutex.RUnlock()
	nullifyObsoleteValues(&memBusLocations)

	busLocationsBytes, err := json.Marshal(memBusLocations)
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}

	stringBusLocations := string(busLocationsBytes)

	response := api.GetBusLocationsResponse{
		Code:      http.StatusOK,
		Locations: stringBusLocations,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}
}

func nullifyObsoleteValues(busLocations *[]tools.BusLocation) {
	for i := range *busLocations {
		(*busLocations)[i].DriverNumber = ""
		(*busLocations)[i].Timestamp = time.Time{}
		(*busLocations)[i].Unknown1 = 0
		(*busLocations)[i].Unknown2 = ""
	}
}
