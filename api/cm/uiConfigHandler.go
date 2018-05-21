/*
###############################################################################
# Licensed Materials - Property of IBM Copyright IBM Corporation 2017, 2018. All Rights Reserved.
# U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP
# Schedule Contract with IBM Corp.
#
# Contributors:
#  IBM Corporation - initial API and implementation
###############################################################################
*/
package cm

import (
	"net/http"
	"regexp"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.ibm.com/IBMPrivateCloud/commands-runner/api/cm/logger"
	"github.ibm.com/IBMPrivateCloud/commands-runner/api/cm/uiConfigManager"
)

//handle BMXCOnfig rest api requests
func handleUIConfig(w http.ResponseWriter, req *http.Request) {
	log.Debug("Entering in handleUIConfig")
	switch req.Method {
	case "GET":
		getUIConfigEndpoint(w, req)
	}
}

/*
Retrieve all Status
URL: /cm/v1/uiconfig/
Method: GET
*/
func getUIConfigEndpoint(w http.ResponseWriter, req *http.Request) {
	log.Debug("Entering in getUIConfigEndpoint")
	//Check format
	validatePath := regexp.MustCompile("/cm/v1/(uiconfig)/([a-z,A-Z,0-9,-]*)$")
	params := validatePath.FindStringSubmatch(req.URL.Path)
	log.Debugf("params=%s", params)
	log.Debug("params size:" + strconv.Itoa(len(params)))
	if len(params) < 3 {
		logger.AddCallerField().Error("Configuration name not found")
		http.Error(w, "Configuration name not found", http.StatusBadRequest)
		return
	}
	//Retrieve the property name
	config, err := uiConfigManager.GetUIConfig(params[2])
	if err == nil {
		w.Write(config)
	} else {
		logger.AddCallerField().Error(err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
	}
}
