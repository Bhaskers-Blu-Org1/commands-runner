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
package extension

import (
	"archive/zip"
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
)

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

func createFileUploadRequest(pathToFile, extensionName string, t *testing.T) *http.Request {
	var req *http.Request
	if pathToFile != "" {
		zipit("../../test/resource/extensions/custom-extension", pathToFile)
		body, _ := os.Open(pathToFile)
		writer := multipart.NewWriter(body)
		req, _ = http.NewRequest("POST", "/cr/v1/extension/action=register", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Content-Disposition", "upload; filename="+filepath.Base(pathToFile))
	} else {
		req, _ = http.NewRequest("POST", "/cr/v1/extension/action=register", nil)
	}
	req.Header.Set("Extension-Name", extensionName)
	return req
}

func TestRegisterExistingExtension(t *testing.T) {
	t.Log("Entering........... TestRegisterExistingExtension")
	// Setup unit test file structure
	SetExtensionPath("../../test/resource/tmp/")

	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	//	SetExtensionPath("../../test/data/extensions/")
	extensionName := "dummy-extension"
	filename := "dummy-extension.zip"
	if _, err := os.Stat(GetExtensionPath()); os.IsNotExist(err) {
		err := os.Mkdir(GetExtensionPath(), 0777)
		if err != nil {
			t.Error(err.Error())
		}
	}
	if _, err := os.Stat(GetExtensionPathCustom()); os.IsNotExist(err) {
		err = os.Mkdir(GetExtensionPathCustom(), 0777)
		if err != nil {
			t.Error(err.Error())
		}
	}
	if _, err := os.Stat(filepath.Join(GetExtensionPathCustom(), extensionName)); os.IsNotExist(err) {
		err = os.Mkdir(filepath.Join(GetExtensionPathCustom(), extensionName), 0777)
		if err != nil {
			t.Error(err.Error())
		}
	}
	fileCreated, err := os.Create(filepath.Join(GetExtensionPathCustom(), filename))
	if err != nil {
		t.Fatal(err)
	}

	fileCreated.Close()

	// Create and handle request for unit test
	req := createFileUploadRequest("../../test/resource/"+filename, extensionName, t)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleExtension)
	handler.ServeHTTP(rr, req)

	assert("Extension "+extensionName+" already registered\n", rr.Body.String(), t)

	cleanup()
}

func TestRegisterNonExistingExtension(t *testing.T) {
	t.Log("Entering........... TestRegisterNonExistingExtension")

	//Setup filesystem
	SetExtensionPath("../../test/resource/tmp/")
	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	filename := "dummy-extension.zip"
	_ = os.Mkdir(GetExtensionPath(), 0777)
	_ = os.Mkdir(GetExtensionPathCustom(), 0777)

	// Create and Handle request
	req := createFileUploadRequest("../../test/resource/"+filename, "dummy-extension", t)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleExtension)
	handler.ServeHTTP(rr, req)

	assert("Extension registration complete", rr.Body.String(), t)

	if _, err := os.Stat(filepath.Join(GetExtensionPathCustom(), "dummy-extension")); os.IsNotExist(err) {
		t.Errorf("project was not unzipped %v\n", err)
	}

	cleanup()
}

func TestRegisterCustomExtension(t *testing.T) {
	t.Log("Entering........... TestExtensionUnzip")
	// Dummy GetExtensionPath()
	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	SetExtensionPath("../../test/resource/tmp/")
	filename := "dummy-extension.zip"
	extensionName := "blahblahblah"
	_ = os.Mkdir(GetExtensionPath(), 0777)
	_ = os.Mkdir(GetExtensionPathCustom(), 0777)

	req := createFileUploadRequest("../../test/resource/"+filename, extensionName, t)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleExtension)
	handler.ServeHTTP(rr, req)

	assert("Extension registration complete", rr.Body.String(), t)

	path := filepath.Join(GetExtensionPathCustom(), extensionName)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("The path: %s, does not exist", path)
	}

	path = filepath.Join(path, "extension-manifest.yml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("The path: %s, does not exist", path)
	}

	path = filepath.Join(path, "/scripts/success.sh")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("The path: %s, does not exist", path)
	}

	//	cleanup()
}

func TestRegisterCustomExtensionWithIBMExtensionName(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	t.Log("Entering........... TestRegisterCustomExtensionWithIBMExtensionName")
	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	SetExtensionPath("../../test/resource/tmp/")

	//Setup filesystem
	SetExtensionPath("../../test/resource/tmp/")
	filename := "dummy-extension.zip"
	extensionName := "cfp-ext-template"
	_ = os.Mkdir(GetExtensionPath(), 0777)
	_ = os.Mkdir(GetExtensionPathCustom(), 0777)

	// Create and Handle request
	req := createFileUploadRequest("../../test/resource/"+filename, extensionName, t)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleExtension)
	handler.ServeHTTP(rr, req)

	assert("Extension name is already used by "+EmbeddedExtensions+" extension\n", rr.Body.String(), t)
	cleanup()
}

func TestRegisterIBMExtension(t *testing.T) {
	t.Log("Entering........... TestRegisterIBMExtension")
	SetExtensionPath("../../test/resource/tmp/")
	SetEmbeddedExtensionsRepositoryPath("../../test/repo_local/")

	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	SetExtensionPath("../../test/resource/tmp/")
	extensionName := "cfp-ext-template"
	_ = os.Mkdir(GetExtensionPath(), 0777)
	_ = os.Mkdir(GetExtensionPathCustom(), 0777)

	// Create and Handle request
	req := createFileUploadRequest("", extensionName, t)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleExtension)
	handler.ServeHTTP(rr, req)

	assert("Extension registration complete", rr.Body.String(), t)
	cleanup()
}

func TestRegisterIBMExtensionFilesExists(t *testing.T) {
	t.Log("Entering........... TestRegisterIBMExtensionFilesExists")
	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	SetExtensionPath("../../test/resource/tmp/")
	SetEmbeddedExtensionsRepositoryPath("../../test/repo_local/")
	extensionName := "cfp-ext-template"
	_ = os.Mkdir(GetExtensionPath(), 0777)
	_ = os.Mkdir(GetExtensionPathEmbedded(), 0777)

	// Create and Handle request
	req := createFileUploadRequest("", extensionName, t)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleExtension)
	handler.ServeHTTP(rr, req)

	assert("Extension registration complete", rr.Body.String(), t)

	path := filepath.Join(GetExtensionPathEmbedded(), extensionName)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("The path: %s, does not exist", path)
	}
	cleanup()
}

func TestDeletionEndpointExists(t *testing.T) {
	t.Log("Entering........... TestExtensionDeletion")
	SetExtensionPath("../../test/resource/tmp/")
	extensionName := "dummy-extension"
	_ = os.Mkdir(GetExtensionPath(), 0777)
	_ = os.Mkdir(GetExtensionPathCustom(), 0777)
	_ = os.Mkdir(filepath.Join(GetExtensionPathCustom(), extensionName), 0777)

	req, err := http.NewRequest("DELETE", "/cr/v1/extension/action=unregister?name="+extensionName, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleExtension)
	handler.ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Fatalf("Delete returned: %v", rr.Code)
	}

	cleanup()
}

func TestDeletionExtensionExists(t *testing.T) {
	t.Log("Entering........... TestDeletionExtensionExists")
	SetExtensionPath("../../test/resource/tmp/")
	extensionName := "dummy-extension2"
	_ = os.Mkdir(GetExtensionPath(), 0777)
	_ = os.Mkdir(GetExtensionPathCustom(), 0777)
	_ = os.Mkdir(GetExtensionPathCustom()+"/dummy-extension", 0777)

	req, err := http.NewRequest("DELETE", "/cr/v1/extension/action=unregister?name="+extensionName, nil)
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
	SetExtensionPath("../../test/resource/tmp/")
	extensionName := "dummy-extension"
	dontDeleteFile := "do-not-delete.zip"
	deleteFile := "dummy-extension.zip"
	err := os.Mkdir(GetExtensionPath(), 0777)
	if err != nil {
		t.Log(err)
	}
	err = os.Mkdir(GetExtensionPathCustom(), 0777)
	if err != nil {
		t.Log(err)
	}
	err = os.Mkdir(GetExtensionPathCustom()+"/dummy-extension", 0777)
	if err != nil {
		t.Log(err)
	}
	os.Create(GetExtensionPathCustom() + dontDeleteFile)
	os.Create(GetExtensionPathCustom() + "/dummy-extension/" + deleteFile)

	req, err := http.NewRequest("DELETE", "/cr/v1/extension/action=unregister?name="+extensionName, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleExtension)
	handler.ServeHTTP(rr, req)

	file, err := os.Stat(GetExtensionPathCustom() + extensionName)
	if file != nil {
		t.Errorf("The extension, %s, was not deleted", extensionName)
	}
	file, err = os.Stat(GetExtensionPathCustom() + deleteFile)
	if file != nil {
		t.Errorf("The extension, %s, was not deleted", extensionName)
	}
	file, err = os.Stat(GetExtensionPathCustom() + dontDeleteFile)
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
	SetExtensionEmbeddedFile("../../test/resource/extensions/test-extensions.txt")
	extensions := [4]string{"dummy-extension1", "dummy-extension2", "dummy-extension3", "dummy-extension4"}
	extensionsIBM := [4]string{"IBM-extension1", "IBM-extension2"}
	SetExtensionPath("../../test/resource/tmp/")
	dontDeleteFile := "do-not-delete.zip"
	deleteFile := "dummy-extension.zip"
	_ = os.Mkdir(GetExtensionPath(), 0777)
	_ = os.Mkdir(GetExtensionPathCustom(), 0777)
	_ = os.Mkdir(GetExtensionPathEmbedded(), 0777)
	for _, extension := range extensions {
		_ = os.Mkdir(filepath.Join(GetExtensionPathCustom(), extension), 0777)
		os.Create(filepath.Join(GetExtensionPathCustom(), extension, "extension-manifest.yml"))
	}
	for _, extension := range extensionsIBM {
		_ = os.Mkdir(filepath.Join(GetExtensionPathEmbedded(), extension), 0777)
		os.Create(filepath.Join(GetExtensionPathEmbedded(), extension, "extension-manifest.yml"))
	}
	os.Create(filepath.Join(GetExtensionPathCustom(), dontDeleteFile))
	os.Create(filepath.Join(GetExtensionPathCustom(), deleteFile))
	os.Mkdir(filepath.Join(GetExtensionPathCustom(), extensions[0], "/do-not-return-embedded-dir"), 0777)
}

func TestListAllExensions(t *testing.T) {
	log.SetLevel(log.DebugLevel)
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
		Name: "dummy-extension1",
		Type: CustomExtensions,
	}
	extensions.Extensions[extension1.Name] = *extension1
	extension2 := &Extension{
		Name: "dummy-extension2",
		Type: CustomExtensions,
	}
	extensions.Extensions[extension2.Name] = *extension2
	extension3 := &Extension{
		Name: "dummy-extension3",
		Type: CustomExtensions,
	}
	extensions.Extensions[extension3.Name] = *extension3
	extension4 := &Extension{
		Name: "dummy-extension4",
		Type: CustomExtensions,
	}
	extensions.Extensions[extension4.Name] = *extension4
	extension5 := &Extension{
		Name: "IBM-extension1",
		Type: EmbeddedExtensions,
	}
	extensions.Extensions[extension5.Name] = *extension5
	extension6 := &Extension{
		Name: "IBM-extension2",
		Type: EmbeddedExtensions,
	}
	extensions.Extensions[extension6.Name] = *extension6
	expected, _ := json.MarshalIndent(&extensions, "", "  ")
	t.Log(rr.Body.String())
	//	expected := `{"extensions":{"extensionsIBM": ["IBM-extension1", "IBM-extension2"],"extensionsCustom": ["dummy-extension1", "dummy-extension2", "dummy-extension3", "dummy-extension4"]}}`
	assert(strings.TrimSpace(string(expected)), strings.TrimSpace(rr.Body.String()), t)
	//assert(expected, rr.Body.String(), t)
	cleanup()
}

func TestListCustomerExensionsWithEmbeddedFolders(t *testing.T) {
	t.Log("TESTING..................... TestListCustomerExensionsWithEmbeddedFolders")
	log.SetLevel(log.DebugLevel)
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
		Name: "dummy-extension1",
		Type: CustomExtensions,
	}
	extensions.Extensions[extension1.Name] = *extension1
	extension2 := &Extension{
		Name: "dummy-extension2",
		Type: CustomExtensions,
	}
	extensions.Extensions[extension2.Name] = *extension2
	extension3 := &Extension{
		Name: "dummy-extension3",
		Type: CustomExtensions,
	}
	extensions.Extensions[extension3.Name] = *extension3
	extension4 := &Extension{
		Name: "dummy-extension4",
		Type: CustomExtensions,
	}
	extensions.Extensions[extension4.Name] = *extension4
	expected, _ := json.MarshalIndent(&extensions, "", "  ")

	//	expected := `{"extensions":{"extensionsCustom": ["dummy-extension1", "dummy-extension2", "dummy-extension3", "dummy-extension4"]}}`
	assert(strings.TrimSpace(string(expected)), strings.TrimSpace(rr.Body.String()), t)
	//assert(expected, rr.Body.String(), t)
	cleanup()
}

func TestListIBMExensions(t *testing.T) {
	t.Log("TESTING..................... TestListIBMExensions")
	setupFileStructureLists()
	SetExtensionPath("../../test/resource/tmp/")

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
		Name: "cfp-ext-template",
		Type: EmbeddedExtensions,
	}
	extension2 := &Extension{
		Name: "cfp-ext-template-auto-location",
		Type: EmbeddedExtensions,
	}

	extensions.Extensions[extension1.Name] = *extension1
	extensions.Extensions[extension2.Name] = *extension2
	expected, _ := json.MarshalIndent(&extensions, "", "  ")
	//	expected := `{"extensions":{"extensionsIBM": ["IBM-extension1", "IBM-extension2"]}}`
	log.Debug(expected)
	log.Debug([]byte(rr.Body.String()))
	assert(strings.TrimSpace(string(expected)), strings.TrimSpace(rr.Body.String()), t)
	cleanup()
}
