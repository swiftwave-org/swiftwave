package cmd

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Swiftwave to the latest minor patch version",
	Long:  `Update Swiftwave to the latest minor patch version`,
	Run: func(cmd *cobra.Command, args []string) {
		swiftwaveVersion := strings.TrimSpace(swiftwaveVersion)
		if swiftwaveVersion == "develop" {
			printError("You should use a stable version of Swiftwave to avail this feature")
			return
		}
		// fetch latest tag commit
		latestTag, err := fetchLatestTag(swiftwaveVersion)
		if err != nil {
			printError("Failed to fetch latest tag")
			return
		}
		if strings.Compare(latestTag, swiftwaveVersion) == 0 {
			printSuccess("Swiftwave is already up-to-date")
			return
		}
		// fetch package download url
		downloadUrl, err := fetchPackageDownloadURL(latestTag)
		if err != nil {
			printError("Failed to fetch package download url")
			return
		}
		// download the package in tmpfs and extract it
		downloadedPackagePath, err := downloadPackage(downloadUrl)
		if err != nil {
			printError("Failed to download and extract the package")
			return
		}
		// extract the package
		err = extractTarGz(downloadedPackagePath, "/tmp/swiftwave-update")
		if err != nil {
			printError("Failed to extract the package")
			return
		}
		// new swiftwave binary path
		newBinaryPath := "/tmp/swiftwave-update/swiftwave"
		// check if new binary exists
		if _, err := os.Stat(newBinaryPath); os.IsNotExist(err) {
			printError("Failed to find new binary")
			return
		}
		// make new binary executable
		err = os.Chmod(newBinaryPath, 0755)
		if err != nil {
			printError("Failed to make new binary executable")
			return
		}
		// replace it at /usr/bin/swiftwave
		err = os.Rename(newBinaryPath, "/usr/bin/swiftwave")
		if err != nil {
			printError("Failed to replace binary")
			return
		}
		// daemon-reload
		runCommand := exec.Command("systemctl", "daemon-reload")
		err = runCommand.Run()
		if err != nil {
			printError("Failed to reload systemd daemon")
			return
		}
		// restart swiftwave service
		runCommand = exec.Command("systemctl", "restart", "swiftwave.service")
		err = runCommand.Run()
		if err != nil {
			printError("Failed to restart swiftwave service")
			return
		}
		printSuccess("Updated Swiftwave to " + latestTag)
	},
}

func getMajorVersion(version string) string {
	splitVersion := strings.Split(version, ".")
	return splitVersion[0]
}

func fetchLatestTag(currentVersion string) (string, error) {
	tagPrefix := getMajorVersion(currentVersion) + "."
	tagListUrl := "https://api.github.com/repos/swiftwave-org/swiftwave/git/refs/tags/" + tagPrefix
	res, err := http.Get(tagListUrl)
	if err != nil || res.StatusCode != 200 {
		return "", err
	}
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	// fetch last tag
	lastTag := ""
	for _, tag := range strings.Split(string(resBody), "refs/tags/") {
		if tag != "" {
			lastTag = strings.Split(tag, "\"")[0]
		}
	}
	if lastTag == "" {
		return "", err
	}
	return lastTag, nil
}

func fetchPackageDownloadURL(tag string) (string, error) {
	packageFileName := "swiftwave-" + tag + "-" + runtime.GOOS + "-" + runtime.GOARCH + ".tar.gz"
	packageDownloadUrl := ""
	// fetch release info
	releaseInfoUrl := "https://api.github.com/repos/swiftwave-org/swiftwave/releases/tags/" + tag
	res, err := http.Get(releaseInfoUrl)
	if err != nil || res.StatusCode != 200 {
		return "", err
	}
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	// fetch download url
	for _, asset := range strings.Split(string(resBody), "browser_download_url") {
		if asset != "" && len(strings.Split(asset, "\"")) > 2 {
			assetUrl := strings.Split(asset, "\"")[2]
			if strings.Contains(assetUrl, packageFileName) &&
				!strings.Contains(assetUrl, packageFileName+".md5") {
				packageDownloadUrl = assetUrl
			}
		}
	}
	if packageDownloadUrl == "" {
		return "", errors.New("failed to fetch package download url")
	}
	return packageDownloadUrl, nil
}

func downloadPackage(url string) (string, error) {
	filename := url[strings.LastIndex(url, "/")+1:]
	// download the package
	file, err := os.Create("/tmp/" + filename)
	if err != nil {
		return "", err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			return
		}
	}(file)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer func(resp *http.Response) {
		err := resp.Body.Close()
		if err != nil {
			return
		}
	}(resp)
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", err
	}
	return "/tmp/" + filename, nil
}

func extractTarGz(downloadedPackagePath, destFolder string) error {
	_ = os.RemoveAll(destFolder)
	err := os.MkdirAll(destFolder, 0755)
	if err != nil {
		return err
	}
	// open downloaded package
	file, err := os.Open(downloadedPackagePath)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			return
		}
	}(file)
	// read gzip file
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer func(gzipReader *gzip.Reader) {
		err := gzipReader.Close()
		if err != nil {
			return
		}
	}(gzipReader)
	// extract tar file
	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		// get filename from header
		filename := header.Name
		// check if file is directory
		if header.Typeflag == tar.TypeDir {
			err = os.MkdirAll(destFolder+"/"+filename, 0755)
			if err != nil {
				return err
			}
			continue
		}
		// create file
		file, err := os.Create(destFolder + "/" + filename)
		if err != nil {
			return err
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				return
			}
		}(file)
		// copy file data
		_, err = io.Copy(file, tarReader)
		if err != nil {
			return err
		}
	}
	return nil
}
