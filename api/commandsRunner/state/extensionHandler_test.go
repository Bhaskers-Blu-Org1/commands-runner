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
package state

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/IBM/commands-runner/api/commandsRunner/global"
)

func init() {
	//log.SetLevel(log.DebugLevel)
}

func cleanup() {
	_ = os.RemoveAll("../../test/resource/tmp")
	_ = os.Remove("../../test/resource/tmp")
}

func assert(expected, actual string, t *testing.T) {
	if actual != expected {
		t.Errorf("expected \n%v actual \n%v", expected, actual)
	}
}

func zipit(source, target string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if baseDir != "" {
			header.Name = filepath.Join(strings.TrimPrefix(path, source))
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})

	return err
}

// func createFileUploadRequest(pathToFile, extensionName string, t *testing.T) *http.Request {
// 	var req *http.Request
// 	if pathToFile != "" {
// 		zipit("../../test/resource/extensions/custom-extension", pathToFile)
// 		body, _ := os.Open(pathToFile)
// 		writer := multipart.NewWriter(body)
// 		req, _ = http.NewRequest("POST", "/cr/v1/extension?extension-name="+extensionName, body)
// 		req.Header.Set("Content-Type", writer.FormDataContentType())
// 		//		req.Header.Set("Content-Disposition", "upload; filename="+filepath.Base(pathToFile))
// 	} else {
// 		req, _ = http.NewRequest("POST", "/cr/v1/extension?extension-name="+extensionName, nil)
// 	}
// 	return req
// }

func createFileFormDataRequest(pathToFile string, extensionName string, t *testing.T) (*http.Request, error) {
	var req *http.Request
	if pathToFile != "" {
		zipit("../../test/resource/extensions/custom-extension", pathToFile)
		file, err := os.Open(pathToFile)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("extension", filepath.Base(pathToFile))
		if err != nil {
			return nil, err
		}
		_, err = io.Copy(part, file)

		// for key, val := range params {
		// 	_ = writer.WriteField(key, val)
		// }
		err = writer.Close()
		if err != nil {
			return nil, err
		}

		var extensionAttribute string
		if extensionName != "" {
			extensionAttribute = "?extension-name=" + extensionName
		}
		log.Debug("extensionAttribute:" + extensionAttribute)
		req, err := http.NewRequest("POST", "/cr/v1/extension"+extensionAttribute, body)
		// t.Logf("req: %v", req)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		return req, err
	} else {
		var extensionAttribute string
		if extensionName != "" {
			extensionAttribute = "?extension-name=" + extensionName
		}
		req, _ = http.NewRequest("POST", "/cr/v1/extension"+extensionAttribute, nil)
	}
	//	req.Header.Set("Extension-Name", extensionName)
	return req, nil
}

func TestRegisterExistingExtension(t *testing.T) {
	t.Log("Entering........... TestRegisterExistingExtension")
	// Setup unit test file structure
	SetExtensionsPath("../../test/resource/tmp/")

	SetExtensionsEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	//	SetExtensionPath("../../test/data/extensions/")
	extensionName := "dummy-extension"
	filename := "dummy-extension.zip"
	if _, err := os.Stat(GetExtensionsPath()); os.IsNotExist(err) {
		err := os.Mkdir(GetExtensionsPath(), 0777)
		if err != nil {
			t.Error(err.Error())
		}
	}
	if _, err := os.Stat(GetExtensionsPathCustom()); os.IsNotExist(err) {
		err = os.Mkdir(GetExtensionsPathCustom(), 0777)
		if err != nil {
			t.Error(err.Error())
		}
	}
	if _, err := os.Stat(filepath.Join(GetExtensionsPathCustom(), extensionName)); os.IsNotExist(err) {
		err = os.Mkdir(filepath.Join(GetExtensionsPathCustom(), extensionName), 0777)
		if err != nil {
			t.Error(err.Error())
		}
	}
	fileCreated, err := os.Create(filepath.Join(GetExtensionsPathCustom(), filename))
	if err != nil {
		t.Fatal(err)
	}

	fileCreated.Close()

	// Create and handle request for unit test
	req, err := createFileFormDataRequest("../../test/resource/"+filename, extensionName, t)
	if err != nil {
		t.Error(err.Error())
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleExtension)
	handler.ServeHTTP(rr, req)

	assert("Extension "+extensionName+" already registered\n", rr.Body.String(), t)

	cleanup()
}

func TestRegisterNonExistingExtension(t *testing.T) {
	t.Log("Entering........... TestRegisterNonExistingExtension")

	//Setup filesystem
	SetExtensionsPath("../../test/resource/tmp/")
	SetExtensionsEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	filename := "dummy-extension.zip"
	_ = os.Mkdir(GetExtensionsPath(), 0777)
	_ = os.Mkdir(GetExtensionsPathCustom(), 0777)

	// Create and Handle request
	req, err := createFileFormDataRequest("../../test/resource/"+filename, "dummy-extension", t)
	if err != nil {
		t.Error(err.Error())
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleExtension)
	handler.ServeHTTP(rr, req)

	assert("Extension registration complete", rr.Body.String(), t)

	if _, err := os.Stat(filepath.Join(GetExtensionsPathCustom(), "dummy-extension")); os.IsNotExist(err) {
		t.Errorf("project was not unzipped %v\n", err)
	}

	cleanup()
}

func TestRegisterCustomExtension(t *testing.T) {
	t.Log("Entering........... TestExtensionUnzip")
	cleanup()
	//log.SetLevel(log.DebugLevel)
	// Dummy GetExtensionPath()
	SetExtensionsEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	SetExtensionsPath("../../test/resource/tmp/")
	filename := "dummy-extension.zip"
	extensionName := "blahblahblah"
	_ = os.Mkdir(GetExtensionsPath(), 0777)
	_ = os.Mkdir(GetExtensionsPathCustom(), 0777)

	req, err := createFileFormDataRequest("../../test/resource/"+filename, extensionName, t)
	if err != nil {
		t.Error(err.Error())
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleExtension)
	handler.ServeHTTP(rr, req)

	assert("Extension registration complete", rr.Body.String(), t)

	path := filepath.Join(GetExtensionsPathCustom(), extensionName)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("The path: %s, does not exist", path)
	}

	path = filepath.Join(path, global.DefaultExtenstionManifestFile)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("The path: %s, does not exist", path)
	}

	path = filepath.Join(path, "/scripts/success.sh")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("The path: %s, does not exist", path)
	}

	cleanup()
}

func TestRegisterCustomExtensionWihtFormData(t *testing.T) {
	t.Log("Entering........... TestExtensionUnzip")
	cleanup()
	//log.SetLevel(log.DebugLevel)
	// Dummy GetExtensionPath()
	SetExtensionsEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	SetExtensionsPath("../../test/resource/tmp/")
	filename := "dummy-extension.zip"
	extensionName := "dummy-extension"
	_ = os.Mkdir(GetExtensionsPath(), 0777)
	_ = os.Mkdir(GetExtensionsPathCustom(), 0777)

	req, err := createFileFormDataRequest("../../test/resource/"+filename, extensionName, t)
	if err != nil {
		t.Error(err.Error())
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleExtension)
	handler.ServeHTTP(rr, req)

	assert("Extension registration complete", rr.Body.String(), t)

	path := filepath.Join(GetExtensionsPathCustom(), extensionName)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("The path: %s, does not exist", path)
	}

	path = filepath.Join(path, global.DefaultExtenstionManifestFile)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("The path: %s, does not exist", path)
	}

	path = filepath.Join(path, "/scripts/success.sh")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("The path: %s, does not exist", path)
	}

	cleanup()
}

func TestRegisterCustomExtensionWithIBMExtensionName(t *testing.T) {
	//	log.SetLevel(log.DebugLevel)
	t.Log("Entering........... TestRegisterCustomExtensionWithIBMExtensionName")
	SetExtensionsEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	SetExtensionsPath("../../test/resource/tmp/")
	filename := "dummy-extension.zip"
	extensionName := "ext-template"
	_ = os.Mkdir(GetExtensionsPath(), 0777)
	_ = os.Mkdir(GetExtensionsPathCustom(), 0777)

	// Create and Handle request
	req, err := createFileFormDataRequest("../../test/resource/"+filename, extensionName, t)
	if err != nil {
		t.Error(err.Error())
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleExtension)
	handler.ServeHTTP(rr, req)

	assert("Error Install:Extension name is already used by embedded extension\n", rr.Body.String(), t)
	cleanup()
}

func TestRegisterIBMExtension(t *testing.T) {
	t.Log("Entering........... TestRegisterIBMExtension")
	SetExtensionsPath("../../test/resource/tmp/")
	SetEmbeddedExtensionsRepositoryPath("../../test/repo_local/")
	SetExtensionsEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	extensionName := "ext-template"
	_ = os.Mkdir(GetExtensionsPath(), 0777)
	_ = os.Mkdir(GetExtensionsPathCustom(), 0777)

	// Create and Handle request
	req, err := createFileFormDataRequest("", extensionName, t)
	if err != nil {
		t.Error(err.Error())
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleExtension)
	handler.ServeHTTP(rr, req)

	assert("Extension registration complete", rr.Body.String(), t)
	cleanup()
}

func TestRegisterIBMExtensionWithVersion(t *testing.T) {
	t.Log("Entering........... TestRegisterIBMExtensionWithVersion")
	SetExtensionsPath("../../test/resource/tmp/")
	SetEmbeddedExtensionsRepositoryPath("../../test/repo_local/")
	SetExtensionsEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	extensionName := "ext-template-v"
	_ = os.Mkdir(GetExtensionsPath(), 0777)
	_ = os.Mkdir(GetExtensionsPathCustom(), 0777)

	// Create and Handle request
	req, err := createFileFormDataRequest("", extensionName, t)
	if err != nil {
		t.Error(err.Error())
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleExtension)
	handler.ServeHTTP(rr, req)

	assert("Extension registration complete", rr.Body.String(), t)
	cleanup()
}

func TestRegisterIBMExtensionFilesExists(t *testing.T) {
	t.Log("Entering........... TestRegisterIBMExtensionFilesExists")
	SetExtensionsEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	SetExtensionsPath("../../test/resource/tmp/")
	SetEmbeddedExtensionsRepositoryPath("../../test/repo_local/")
	extensionName := "ext-template-2"
	_ = os.Mkdir(GetExtensionsPath(), 0777)
	_ = os.Mkdir(GetExtensionsPathEmbedded(), 0777)

	// Create and Handle request
	req, err := createFileFormDataRequest("", extensionName, t)
	if err != nil {
		t.Error(err.Error())
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleExtension)
	handler.ServeHTTP(rr, req)

	assert("Extension registration complete", rr.Body.String(), t)

	path := filepath.Join(GetExtensionsPathEmbedded(), extensionName)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("The path: %s, does not exist", path)
	}
	cleanup()
}

func TestDeletionEndpointExists(t *testing.T) {
	t.Log("Entering........... TestDeletionEndpointExists")
	SetExtensionsPath("../../test/resource/tmp/")
	extensionName := "dummy-extension"
	_ = os.Mkdir(GetExtensionsPath(), 0777)
	_ = os.Mkdir(GetExtensionsPathCustom(), 0777)
	_ = os.Mkdir(filepath.Join(GetExtensionsPathCustom(), extensionName), 0777)

	req, err := http.NewRequest("DELETE", "/cr/v1/extension?extension-name="+extensionName, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleExtension)
	handler.ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Fatalf("Delete returned: %v - %v", rr.Code, rr.Body.String())
	}

	cleanup()
}

func TestDeletionExtensionExists(t *testing.T) {
	t.Log("Entering........... TestDeletionExtensionExists")
	SetExtensionsPath("../../test/resource/tmp/")
	extensionName := "dummy-extension2"
	_ = os.Mkdir(GetExtensionsPath(), 0777)
	_ = os.Mkdir(GetExtensionsPathCustom(), 0777)
	_ = os.Mkdir(GetExtensionsPathCustom()+"/dummy-extension", 0777)

	req, err := http.NewRequest("DELETE", "/cr/v1/extension?extension-name="+extensionName, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleExtension)
	handler.ServeHTTP(rr, req)

	if rr.Code != 500 {
		t.Fatalf("Extension should not exists. Status code: %v", rr.Code)
	}

	cleanup()
}

func TestDeletionFromFileSystem(t *testing.T) {
	t.Log("Entering........... TestDeletionFromFileSystem")
	SetExtensionsPath("../../test/resource/tmp/")
	extensionName := "dummy-extension"
	dontDeleteFile := "do-not-delete.zip"
	deleteFile := "dummy-extension.zip"
	err := os.Mkdir(GetExtensionsPath(), 0777)
	if err != nil {
		t.Log(err)
	}
	err = os.Mkdir(GetExtensionsPathCustom(), 0777)
	if err != nil {
		t.Log(err)
	}
	err = os.Mkdir(GetExtensionsPathCustom()+"/dummy-extension", 0777)
	if err != nil {
		t.Log(err)
	}
	os.Create(GetExtensionsPathCustom() + dontDeleteFile)
	os.Create(GetExtensionsPathCustom() + "/dummy-extension/" + deleteFile)

	req, err := http.NewRequest("DELETE", "/cr/v1/extension?extension-name="+extensionName, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleExtension)
	handler.ServeHTTP(rr, req)

	file, err := os.Stat(GetExtensionsPathCustom() + extensionName)
	if file != nil {
		t.Errorf("The extension, %s, was not deleted", extensionName)
	}
	file, err = os.Stat(GetExtensionsPathCustom() + deleteFile)
	if file != nil {
		t.Errorf("The extension, %s, was not deleted", extensionName)
	}
	file, err = os.Stat(GetExtensionsPathCustom() + dontDeleteFile)
	if file == nil {
		t.Errorf("The extension, %s, was not suppose to be deleted", extensionName)
	}

	cleanup()
}

func TestListEndpointExists(t *testing.T) {
	req, err := http.NewRequest("GET", "/cr/v1/extensions/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleExtension)
	handler.ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Fatalf("GET endpoints returned: %v", rr.Code)
	}
	cleanup()
}

func setupFileStructureLists() {
	SetExtensionsEmbeddedFile("../../test/resource/extensions/test-extensions.yml")
	extensions := [4]string{"dummy-extension1", "dummy-extension2", "dummy-extension3", "dummy-extension4"}
	extensionsIBM := [4]string{"IBM-extension1", "IBM-extension2"}
	SetExtensionsPath("../../test/resource/tmp/")
	dontDeleteFile := "do-not-delete.zip"
	deleteFile := "dummy-extension.zip"
	_ = os.Mkdir(GetExtensionsPath(), 0777)
	_ = os.Mkdir(GetExtensionsPathCustom(), 0777)
	_ = os.Mkdir(GetExtensionsPathEmbedded(), 0777)
	for _, extension := range extensions {
		_ = os.Mkdir(filepath.Join(GetExtensionsPathCustom(), extension), 0777)
		os.Create(filepath.Join(GetExtensionsPathCustom(), extension, global.DefaultExtenstionManifestFile))
	}
	for _, extension := range extensionsIBM {
		_ = os.Mkdir(filepath.Join(GetExtensionsPathEmbedded(), extension), 0777)
		os.Create(filepath.Join(GetExtensionsPathEmbedded(), extension, global.DefaultExtenstionManifestFile))
	}
	os.Create(filepath.Join(GetExtensionsPathCustom(), dontDeleteFile))
	os.Create(filepath.Join(GetExtensionsPathCustom(), deleteFile))
	os.Mkdir(filepath.Join(GetExtensionsPathCustom(), extensions[0], "/do-not-return-embedded-dir"), 0777)
}

func TestListAllExensions(t *testing.T) {
	//log.SetLevel(log.DebugLevel)
	t.Log("TESTING..................... TestListAllExensions")
	setupFileStructureLists()

	req, err := http.NewRequest("GET", "/cr/v1/extensions?catalog=false", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleExtensions)
	handler.ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Fatalf("GET endpoints returned: %v %v", rr.Code, rr.Body.String())
	}

	var extensions Extensions
	extensions.Extensions = make(map[string]Extension)
	extension1 := &Extension{
		Type: CustomExtensions,
	}
	extensions.Extensions["dummy-extension1"] = *extension1
	extension2 := &Extension{
		Type: CustomExtensions,
	}
	extensions.Extensions["dummy-extension2"] = *extension2
	extension3 := &Extension{
		Type: CustomExtensions,
	}
	extensions.Extensions["dummy-extension3"] = *extension3
	extension4 := &Extension{
		Type: CustomExtensions,
	}
	extensions.Extensions["dummy-extension4"] = *extension4
	extension5 := &Extension{
		Type: EmbeddedExtensions,
	}
	extensions.Extensions["IBM-extension1"] = *extension5
	extension6 := &Extension{
		Type: EmbeddedExtensions,
	}
	extensions.Extensions["IBM-extension2"] = *extension6
	expected, _ := json.MarshalIndent(&extensions, "", "  ")
	t.Log(rr.Body.String())
	//	expected := `{"extensions":{"extensionsIBM": ["IBM-extension1", "IBM-extension2"],"extensionsCustom": ["dummy-extension1", "dummy-extension2", "dummy-extension3", "dummy-extension4"]}}`
	assert(strings.TrimSpace(string(expected)), strings.TrimSpace(rr.Body.String()), t)
	//assert(expected, rr.Body.String(), t)
	cleanup()
}

func TestListCustomerExensionsWithEmbeddedFolders(t *testing.T) {
	t.Log("TESTING..................... TestListCustomerExensionsWithEmbeddedFolders")
	//log.SetLevel(log.DebugLevel)
	setupFileStructureLists()

	req, err := http.NewRequest("GET", "/cr/v1/extensions?filter="+CustomExtensions, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleExtensions)
	handler.ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Fatalf("GET endpoints returned: %v", rr.Code)
	}

	var extensions Extensions
	extensions.Extensions = make(map[string]Extension)
	extension1 := &Extension{
		Type: CustomExtensions,
	}
	extensions.Extensions["dummy-extension1"] = *extension1
	extension2 := &Extension{
		Type: CustomExtensions,
	}
	extensions.Extensions["dummy-extension2"] = *extension2
	extension3 := &Extension{
		Type: CustomExtensions,
	}
	extensions.Extensions["dummy-extension3"] = *extension3
	extension4 := &Extension{
		Type: CustomExtensions,
	}
	extensions.Extensions["dummy-extension4"] = *extension4
	expected, _ := json.MarshalIndent(&extensions, "", "  ")

	//	expected := `{"extensions":{"extensionsCustom": ["dummy-extension1", "dummy-extension2", "dummy-extension3", "dummy-extension4"]}}`
	assert(strings.TrimSpace(string(expected)), strings.TrimSpace(rr.Body.String()), t)
	//assert(expected, rr.Body.String(), t)
	cleanup()
}

func TestListIBMExensions(t *testing.T) {
	//log.SetLevel(log.DebugLevel)
	t.Log("TESTING..................... TestListIBMExensions")
	setupFileStructureLists()
	SetExtensionsPath("../../test/resource/tmp/")

	req, err := http.NewRequest("GET", "/cr/v1/extensions?filter="+EmbeddedExtensions+"&catalog=true", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleExtensions)
	handler.ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Fatalf("GET endpoints returned: %v", rr.Code)
	}

	var extensions Extensions
	extensions.Extensions = make(map[string]Extension)
	extension1 := &Extension{
		Type: EmbeddedExtensions,
	}
	extension2 := &Extension{
		Type:    EmbeddedExtensions,
		Version: "1.0.0",
	}
	extension3 := &Extension{
		Type: EmbeddedExtensions,
	}
	extension4 := &Extension{
		Type: EmbeddedExtensions,
	}
	extension5 := &Extension{
		Type: EmbeddedExtensions,
	}
	extension6 := &Extension{
		Type: EmbeddedExtensions,
	}
	extension7 := &Extension{
		Type: EmbeddedExtensions,
	}
	extension8 := &Extension{
		Type: EmbeddedExtensions,
	}
	extension9 := &Extension{
		Type: EmbeddedExtensions,
	}

	extensions.Extensions["ext-template"] = *extension1
	extensions.Extensions["ext-template-v"] = *extension2
	extensions.Extensions["ext-template-auto-location"] = *extension3
	extensions.Extensions["ext-template-2"] = *extension4
	extensions.Extensions["ext-insert-delete"] = *extension5
	extensions.Extensions["ext-insert-delete-by-name"] = *extension6
	extensions.Extensions["ext-template-insert-delete-handler"] = *extension7
	extensions.Extensions["ext-insert-delete-handler"] = *extension8
	extensions.Extensions["ext-template-states-run-success-with-extension"] = *extension9

	expected, _ := json.MarshalIndent(&extensions, "", "  ")
	//	expected := `{"extensions":{"extensionsIBM": ["IBM-extension1", "IBM-extension2"]}}`
	log.Debug(expected)
	log.Debug([]byte(rr.Body.String()))
	assert(strings.TrimSpace(string(expected)), strings.TrimSpace(rr.Body.String()), t)
	cleanup()
}
