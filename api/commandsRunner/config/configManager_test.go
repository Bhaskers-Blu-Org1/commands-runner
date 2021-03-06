/*
################################################################################
# Copyright 2019 IBM Corp. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
################################################################################
*/
package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/IBM/commands-runner/api/commandsRunner/global"
	"github.com/IBM/commands-runner/api/commandsRunner/properties"
	"github.com/IBM/commands-runner/api/commandsRunner/state"
)

const COPYRIGHT_TEST string = `###############################################################################
# Licensed Materials - Property of IBM Copyright IBM Corporation 2017, 2019. All Rights Reserved.
# U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP
# Schedule Contract with IBM Corp.
#
# Contributors:
#  IBM Corporation - initial API and implementation
###############################################################################`

//var bmxConfigString string = "{\"Prop1\":{\"name\":\"Prop1\",\"value\":\"Val1\"},\"Prop2\":{\"name\":\"Prop2\",\"value\":\"Val2\"},\"subnet\":{\"name\":\"subnet\",\"value\":\"192.168.100.0/24\"}}"

var configString string = global.ConfigRootKey + ":\n  env_name: \"itdove\"\n  host_directory: \"/itdove/data\"\n  subnet: \"192.168.100.0/24\""

//var global.ConfigDirectory string = "../../test/resource"
//var properties properties.Properties

func TestSetConfigPath(t *testing.T) {
	t.Log("Entering... TestSetConfigPath")
	global.ConfigDirectory = "../../test/resource"
	os.Remove(global.ConfigDirectory)
	SetConfigPath(global.ConfigDirectory)
}

func TestSetProperties(t *testing.T) {
	t.Log("Entering... TestSetproperties.Properties")
	props = make(properties.Properties)
	global.ConfigDirectory = "../../test/resource"
	extensionPath, err := global.CopyToTemp("TestSetProperties", "../../test/resource/extensions/")
	if err != nil {
		t.Fatal(err)
	}
	state.SetExtensionsPath(extensionPath)
	//	t.Error(global.ConfigDirectory)
	os.MkdirAll(global.ConfigDirectory, 0744)
	props["Prop3"] = "Val3"
	props["Prop4"] = "Val4"
	props["subnet"] = "192.168.100.0/24"
	err = SetProperties("config-manager-test", props)
	if err != nil {
		t.Error(err.Error())
	}
	global.RemoveTemp("TestSetProperties")
}

func TestGetProperties(t *testing.T) {
	t.Log("Entering... TestGetproperties.Properties")
	t.Logf("%s\n", configString)
	global.ConfigDirectory = "../../test/resource"
	extensionPath, err := global.CopyToTemp("TestGetProperties", "../../test/resource/extensions/")
	if err != nil {
		t.Fatal(err)
	}
	state.SetExtensionsPath(extensionPath)
	dataDirectory := state.GetRootExtensionPath("../../test/resource/extensions", "config-manager-test")
	t.Log("dataDirectory:" + dataDirectory)
	err = ioutil.WriteFile(filepath.Join(dataDirectory, global.ConfigYamlFileName), []byte(configString), 0644)
	if err != nil {
		t.Error("Can not create temp file")
	}
	SetConfigPath(global.ConfigDirectory)
	//t.Log(properties)
	propertiesAux, err := GetProperties("config-manager-test")
	t.Logf("%s\n", propertiesAux)
	if err != nil {
		t.Error(err.Error())
	}
	t.Logf("%s\n", propertiesAux["env_name"])
	//p, err := FindProperty("Prop1")
	val, err := properties.GetValueAsString(propertiesAux, "env_name")
	if err != nil {
		t.Error(err.Error())
	}
	if val != "itdove" {
		t.Error("Expected value Val1 and get" + val)
	}
	global.RemoveTemp("TestGetProperties")
}

func TestFindProperty(t *testing.T) {
	t.Log("Entering... TestFindProperty")
	t.Logf("%s\n", configString)
	global.ConfigDirectory = "../../test/resource"
	extensionPath, err := global.CopyToTemp("TestFindProperty", "../../test/resource/extensions/")
	if err != nil {
		t.Fatal(err)
	}
	state.SetExtensionsPath(extensionPath)
	dataDirectory := state.GetRootExtensionPath("../../test/resource/extensions", "config-manager-test")
	err = ioutil.WriteFile(filepath.Join(dataDirectory, global.ConfigYamlFileName), []byte(configString), 0644)
	if err != nil {
		t.Error("Can not create temp file")
	}
	SetConfigPath(global.ConfigDirectory)
	p, err := FindProperty("config-manager-test", "env_name")
	if err != nil {
		t.Error(err.Error())
	}
	if p == nil {
		t.Error("Can not retreive properties")
	}
	t.Log(p)
	if val, ok := p["value"].(string); ok {
		if val != "itdove" {
			t.Error("Expected value Val1 and get" + val)
		}
	} else {
		t.Error("Not a string")
	}
	p, err = FindProperty("config-manager-test", "Prop3")
	if err == nil {
		t.Error("Expected not found and found")
	}
	global.RemoveTemp("TestFindProperty")
}

func TestRemoveProperty(t *testing.T) {
	t.Log("Entering... TestRemoveProperty")
	t.Logf("%s\n", configString)
	global.ConfigDirectory = "../../test/resource"
	extensionPath, err := global.CopyToTemp("TestRemoveProperty", "../../test/resource/extensions/")
	if err != nil {
		t.Fatal(err)
	}
	state.SetExtensionsPath(extensionPath)
	dataDirectory := state.GetRootExtensionPath("../../test/resource/extensions", "config-manager-test")
	err = ioutil.WriteFile(filepath.Join(dataDirectory, global.ConfigYamlFileName), []byte(configString), 0644)
	if err != nil {
		t.Error("Can not create temp file")
	}
	SetConfigPath(global.ConfigDirectory)
	err = RemoveProperty("config-manager-test", "Prop1")
	if err != nil {
		t.Error(err.Error())
	}
	p, err := FindProperty("config-manager-test", "Prop1")
	if p != nil {
		t.Error("Expected not found and found")
	}
	global.RemoveTemp("TestRemoveProperty")
}
