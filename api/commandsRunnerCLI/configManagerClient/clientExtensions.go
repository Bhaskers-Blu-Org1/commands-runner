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
package configManagerClient

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/extensionManager"
	"github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/global"
)

func (cmc *ConfigManagerClient) getExtensions(extensionsToList string, catalog bool) (string, error) {
	url := "/extensions/?filter=" + extensionsToList + ";amp&catalog=" + strconv.FormatBool(catalog)
	data, errCode, err := cmc.RestCall(http.MethodGet, global.BaseURL, url, nil, nil)
	if errCode != http.StatusOK {
		return "", errors.New("Unable to get extensions, please check logs")
	}
	return data, err
}

func (cmc *ConfigManagerClient) GetExtensions(extensionToList string, catalog bool) (string, error) {
	data, err := cmc.getExtensions(extensionToList, catalog)
	if err != nil {
		return "", err
	}
	out := ""
	if cmc.OutputFormat == "text" {
		var extensions extensionManager.Extensions
		jsonErr := json.Unmarshal([]byte(data), &extensions)
		if jsonErr != nil {
			return "", jsonErr
		}
		for _, v := range extensions.Extensions {
			out += fmt.Sprintf("=>\n")
			out += fmt.Sprintf("name : %s\n", v.Name)
			out += fmt.Sprintf("type : %s\n", v.Type)
		}
		return out, nil
	}
	return convertJSONToYAML(data)
}
